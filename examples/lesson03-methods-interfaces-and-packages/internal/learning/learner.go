package learning

// Learner 表示正在学习 Go 后端的学习者。
type Learner struct {
	Name             string
	CompletedLessons int
}

// Level 根据完成课程数计算学习阶段。
// 这个方法不修改 Learner，所以使用值接收者。
func (l Learner) Level() string {
	switch {
	case l.CompletedLessons >= 8:
		return "advanced"
	case l.CompletedLessons >= 3:
		return "intermediate"
	default:
		return "beginner"
	}
}

// CompleteLesson 会修改完成课程数，所以使用指针接收者。
func (l *Learner) CompleteLesson() {
	l.CompletedLessons++
}
