// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

// CleanPath 是 path.Clean 的 URL 版本，它返回 p 的规范 URL 路径，
// 通过消除 . 和 .. 元素。
//
// 以下规则将被反复应用，直到无法进一步处理为止：
//  1. 将多个斜杠替换为单个斜杠。
//  2. 消除每个 . 路径名元素（当前目录）。
//  3. 消除每个内部的 .. 路径名元素（父目录）
//     以及它前面的非 .. 元素。
//  4. 消除开始于根路径的 .. 元素：
//     即，在路径开始处将 "/.." 替换为 "/"。
//
// 如果这个过程的结果是空字符串，则返回 "/"。

func CleanPath(p string) string {
	const stackBufSize = 128

	// 将空字符串转换为 "/"
	if p == "" {
		return "/"
	}

	// 在栈上分配一个合理大小的缓冲区，以避免在常见情况下进行内存分配。
	// 如果需要更大的缓冲区，则会动态分配。
	buf := make([]byte, 0, stackBufSize)

	n := len(p)

	// 不变量：
	//      从路径读取；r 是要处理的下一个字节的索引。
	//      写入到 buf；w 是要写入的下一个字节的索引。

	// 路径必须以 '/' 开头
	r := 1
	w := 1

	if p[0] != '/' {
		r = 0

		if n+1 > stackBufSize {
			buf = make([]byte, n+1)
		} else {
			buf = buf[:n+1]
		}
		buf[0] = '/'
	}

	trailing := n > 1 && p[n-1] == '/'

	// 没有像 path 包那样的 'lazybuf' 可能会显得有些笨拙，但循环
	// 完全内联（bufApp 调用）。
	// 因此，与 path 包相比，这个循环没有昂贵的函数
	// 调用（除了必要时的 make）

	for r < n {
		switch {
		case p[r] == '/':
			// 空路径元素，尾随斜杠将在末尾添加
			r++

		case p[r] == '.' && r+1 == n:
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/':
			// . element
			r += 2

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// .. element: remove to last /
			r += 3

			if w > 1 {
				// can backtrack
				w--

				if len(buf) == 0 {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}

		default:
			// 真实路径元素。
			// 如有必要，添加斜杠
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}

			// Copy element
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	// 重新添加尾部斜杠
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}

	// 如果原始字符串未被修改（或仅在末尾被缩短），
	// 返回原始字符串的相应子字符串。
	// 否则从缓冲区返回一个新字符串。
	if len(buf) == 0 {
		return p[:w]
	}
	return string(buf[:w])
}

// 内部辅助函数，如有必要，可延迟创建缓冲区。
// 对此函数的调用将被内联。
func bufApp(buf *[]byte, s string, w int, c byte) {
	b := *buf
	if len(b) == 0 {
		// 到目前为止，原始字符串未被修改。
		// 如果下一个字符与原始字符串中的字符相同，
		// 我们还不需要分配缓冲区。
		if s[w] == c {
			return
		}

		// 否则使用栈缓冲区（如果其足够大），或者
		// 在堆上分配一个新的缓冲区，并复制所有之前的字符。
		if l := len(s); l > cap(b) {
			*buf = make([]byte, len(s))
		} else {
			*buf = (*buf)[:l]
		}
		b = *buf

		copy(b, s[:w])
	}
	b[w] = c
}
