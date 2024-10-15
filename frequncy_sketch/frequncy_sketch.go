package frequncy_sketch

import (
	"gaffeine/global"
	"gaffeine/utils"
	"math"
	"math/bits"
)

const (
	ResetMask = int64(0x7777777777777777) // uint64类型
	OneMask   = int64(0x1111111111111111) // uint64类型
)

// FrequencySketch maintains a 4-bit CountMinSketch [1] with periodic aging to provide the popularity history for the TinyLfu admission policy [2].
// 本类维护一个 4 位 CountMinSketch [1]，并定期进行衰减，以提供 TinyLfu 接纳策略 [2] 的流行历史。
// The time and space efficiency of the sketch allows it to cheaply estimate the frequency of an entry in a stream of caches access events.
// 该草图的时间和空间效率使其能够廉价地估算在一系列缓存访问事件中的条目频率。
//
// The counter matrix is represented as a single-dimensional array holding 16 counters per slot.
// 计数矩阵表示为一个一维数组，每个槽位持有 16 个计数器。
// A fixed depth of four balances the accuracy and cost, resulting in a width of four times the length of the array.
// 固定的深度为四平衡了准确性和成本，使得数组的宽度是数组长度的四倍。
// To retain an accurate estimation, the array's length equals the maximum number of entries in the caches,
// 为了保持准确的估计，数组的长度等于缓存中的最大条目数，
// increased to the closest power-of-two to exploit more efficient bit masking.
// 增加到最接近的二的幂以利用更高效的位掩码。
// This configuration results in a confidence of 93.75% and an error bound of e / width.
// 该配置的置信度为 93.75%，误差界限为 e / 宽度。
//
// To improve hardware efficiency, an item's counters are constrained to a 64-byte block, which is the size of an L1 caches line.
// 为了提高硬件效率，条目的计数器被限制在 64 字节的块中，这正是 L1 缓存行的大小。
// This differs from the theoretical ideal where counters are uniformly distributed to minimize collisions.
// 这与理论理想状态不同，后者要求计数器均匀分布以最小化碰撞。
// In that configuration, the memory accesses are not predictable and lack spatial locality,
// 在这种配置中，内存访问是不可预测的，并且缺乏空间局部性，
// which may cause the pipeline to need to wait for four memory loads.
// 这可能导致流水线需要等待四次内存加载。
// Instead, the items are uniformly distributed to blocks, and each counter is uniformly selected from a distinct 16-byte segment.
// 相反，条目均匀分布到块中，每个计数器均匀地从不同的 16 字节段中选择。
// While the runtime memory layout may result in the blocks not being caches-aligned,
// 虽然运行时内存布局可能导致块未对齐缓存，
// the L2 spatial prefetcher tries to load aligned pairs of caches lines, so the typical cost is only one memory access.
// 但 L2 空间预取器尝试加载对齐的缓存行对，因此典型成本仅为一次内存访问。
//
// The frequency of all entries is aged periodically using a sampling window based on the maximum number of entries in the caches.
// 所有条目的频率都使用基于缓存中最大条目数的采样窗口定期衰减。
// This is referred to as the reset operation by TinyLfu and keeps the sketch fresh by dividing all counters by two and subtracting based on the number of odd counters found.
// 这被 TinyLfu 称为重置操作，通过将所有计数器除以二并根据找到的奇数计数器的数量进行减法，保持草图的新鲜度。
// The O(n) cost of aging is amortized, ideal for hardware prefetching, and uses inexpensive bit manipulations per array location.
// 衰减的 O(n) 成本是摊销的，非常适合硬件预取，并对每个数组位置使用廉价的位操作。
//
// [1] An Improved Data Stream Summary: The Count-Min Sketch and its Applications
// [1] 改进的数据流摘要：Count-Min Sketch 及其应用
// http://dimacs.rutgers.edu/~graham/pubs/papers/cm-full.pdf
//
// [2] TinyLFU: A Highly Efficient Cache Admission Policy
// [2] TinyLFU：一种高效的缓存接纳策略
// https://dl.acm.org/citation.cfm?id=3149371
//
// [3] Hash Function Prospector: Three round functions
// [3] 哈希函数探测器：三轮函数
// https://github.com/skeeto/hash-prospector#three-round-functions
type FrequencySketch[K global.Key] struct {
	KeyType    K
	SampleSize int // 需要进行Reset的容量
	BlockMask  int // 一个块(8个int64大小）的掩码
	Size       int // 当前已经使用的计数器个数，这个是一个评估值，不是一个精确值
	Table      []int64
}

func New[K global.Key]() *FrequencySketch[K] {
	sketch := FrequencySketch[K]{
		Table:      nil,
		SampleSize: 0,
		BlockMask:  0,
		Size:       0,
	}
	return &sketch
}

// EnsureCapacity Initializes and increases the capacity of this <tt>FrequencySketch</tt> instance, if necessary,
// to ensure that it can accurately estimate the popularity of elements given the maximum size of
// the caches. This operation forgets all previous counts when resizing.
// @param maximumSize the maximum size of the caches
func (f *FrequencySketch[K]) EnsureCapacity(maximumSize int) *FrequencySketch[K] {
	if maximumSize <= 0 {
		maximumSize = 8
	}

	maximum := int(utils.Min(maximumSize, math.MaxInt32>>1))
	if f.Table != nil && len(f.Table) >= maximum {
		return f
	}
	newSize := int(utils.Max(utils.CeilingPowerOfTwo32(maximum), 8))
	f.Table = make([]int64, newSize)
	if maximumSize == 0 {
		f.SampleSize = 10
	} else {
		f.SampleSize = 10 * maximum
	}

	// a）>>3，是因为：64位架构CPU的一个缓存块大小是64个字节。
	// 				 而8个int64为一个块，刚好是64个字节，从而有更快的读取速度
	// b）-1，是因为：len(f.Table)>>3得到的数一定是一个首位是1，其他位是0的数。
	// 				-1后，首位是0，其他位是1，从而得到一个掩码。
	f.BlockMask = len(f.Table)>>3 - 1

	if int32(f.SampleSize) <= 0 { // 防止溢出
		f.SampleSize = math.MaxInt32
	} else if f.SampleSize > math.MaxInt32 {
		f.SampleSize = math.MaxInt32
	}

	f.Size = 0

	return f
}

// Increment Increments the popularity of the element if it does not exceed the maximum (15). The popularity
// of all elements will be periodically down sampled when the observed events exceed a threshold.
// This process provides a frequency aging to allow expired long term entries to fade away.
// @param e the element to add
func (f *FrequencySketch[K]) Increment(key K) *FrequencySketch[K] {
	// 4、5、6、7存放的是table的index
	// 0、1、2、3存放的是table[index]的计数器的offset
	// 注意：table[index]是一个long，所以有64/4=16个计数器
	index := make([]int, 8)
	blockHash := f.spread(hashcode(key))
	counterHash := f.rehash(blockHash)
	block := int(blockHash&uint32(f.BlockMask)) << 3 // 找到table的位置，table的一个块有8个uint64，所以要<<3

	for i := 0; i < 4; i++ {
		h := int(counterHash >> (i << 3))      // i<<3 在循环中，分别是：0、8、16、24
		index[i] = (h >> 1) & 15               // 执行>>1，是为了提高hash的分散性；&15 是把结果控制在0-15之间
		offset := h & 1                        // offset 结果只能是0或者1，换句话说，h & 1 相当于 h % 2的结果，也就是 offset 代表h是奇数还是偶数
		index[i+4] = block + offset + (i << 1) // block是table的下标；offset的结果只能是0或者1；i << 1 只能是0、2、4、6；所以，最后的值：block + 0/1 + 0/2/4/6
	}
	added := f.incrementAt(index[4], index[0])
	added = f.incrementAt(index[5], index[1]) || added
	added = f.incrementAt(index[6], index[2]) || added
	added = f.incrementAt(index[7], index[3]) || added

	if added {
		f.Size += 1
		if f.Size == f.SampleSize {
			f.Reset()
		}
	}
	return f
}

// Frequency Returns the estimated number of occurrences of an element, up to the maximum (15).
// @param e the element to count occurrences of
// @return the estimated number of occurrences of the element; possibly zero but never negative
func (f *FrequencySketch[K]) Frequency(key K) int {
	count := make([]int, 4)
	blockHash := f.spread(hashcode(key))
	counterHash := f.rehash(blockHash)
	block := int(blockHash&uint32(f.BlockMask)) << 3

	for i := 0; i < 4; i++ {
		h := int(counterHash >> (i << 3)) // i<<3 在循环中，分别是：0、8、16、24
		index := (h >> 1) & 15            // 执行>>1，是为了提高hash的分散性；&15 是把结果控制在0-15之间
		offset := h & 1                   // offset 结果只能是0或者1，换句话说，h & 1 相当于 h % 2的结果，也就是 offset 代表h是奇数还是偶数
		tableV := uint64(f.Table[block+offset+(i<<1)])
		count[i] = int(tableV >> (index << 2) & uint64(0xf))
	}

	return int(utils.Min(utils.Min(count[0], count[1]), utils.Min(count[2], count[3])))
}

// spread Applies a supplemental hash function to defend against a poor quality hash.
// https://github.com/skeeto/hash-prospector#three-round-functions
func (f *FrequencySketch[K]) spread(x uint32) uint32 {
	x ^= x >> 17
	x *= 0xed5ad4bb
	x ^= x >> 11
	x *= 0xac4c1b51
	x ^= x >> 15
	return x
}

// rehash Applies another round of hashing for additional randomization.
// https://github.com/skeeto/hash-prospector#three-round-functions
func (f *FrequencySketch[K]) rehash(x uint32) uint32 {
	x *= 0x31848bab
	x ^= x >> 14
	return x
}

// incrementAt Increments the specified counter by 1 if it is not already at the maximum value (15).
// @param i the table index (16 counters if table[i])
// @param j the counter to increment
// @return if incremented
func (f *FrequencySketch[K]) incrementAt(i, j int) bool {

	// 相当于j*4，那么offset的结果是[0, 60]。
	// 这个结果代表了第 j 个计数器在 long 值中的具体位置。1一个计数器是4位。所以：
	//    	j=0，那么offset=0，表示0-3；
	//		j=1，那么offset=4，表示4-7；
	//		j=2，那么offset=8，表示8-11；
	//  	...... 以此类推
	offset := j << 2

	// 0xfL 表示一个值为 1111（4 个二进制 1）的 long 类型常量，也就是一个 4-bit 的掩码。
	// 0xfL << offset 将这个 4-bit 掩码移到相应的位置上，对应到 long 值中某个 4-bit 计数器的位置。
	mask := int64(0xf) << offset

	if (f.Table[i] & mask) != mask {
		// 判断是否已经达到最大数15
		f.Table[i] += int64(1) << offset // 如果是，则将这个计数器值加1; 1L << offset 就是把1左移offset位置，也就是这个计数器的位置
		return true
	}
	return false
}

// Reset reduces every counter by half of its original value.
func (f *FrequencySketch[K]) Reset() *FrequencySketch[K] {
	// count: 表示有多少个有效计数器。
	count := 0
	for i := 0; i < len(f.Table); i++ {
		// 只能统计奇数的计算器，所以，count不是精确值，而是估算值
		// 为什么要用估算，而不用精确算法。是为了性能和效率。
		count += bits.OnesCount64(uint64(f.Table[i] & OneMask))

		// >>>1 相当于除以2，但是注意：1111，1111 执行后，结果为：0111，1111；可以发现第2个计数器仍然是1111，
		// 所以，还需要 & ResetMask，将第2个计数器的高位设置为0，这样 0111，1111 变成了 0111, 0111
		f.Table[i] = int64(uint64(f.Table[i])>>1) & ResetMask
	}
	f.Size = (f.Size - (count >> 2)) >> 1
	return f
}
