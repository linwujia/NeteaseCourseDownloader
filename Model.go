package main

type CourseChapter struct {
	Title    string
	Link     string
	Sessions []*CourseSession
}

//章
type CourseSession struct {
	Title    string
	Children []*Child
}

// 节或者具体的视频
type Child struct {
	Title    string
	Link     string
	Children []*Child
}
