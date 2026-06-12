package main

// Greet 返回一段问候语。
// 这类纯函数不依赖外部状态，最适合作为第一批单元测试对象。
func Greet(name string) string {
	if name == "" {
		name = "friend"
	}

	return "Hello, " + name + "!"
}
