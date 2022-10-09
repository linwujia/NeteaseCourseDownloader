package main

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
	Uuid    interface{} `json:"uuid"`
}

type Course struct {
	Id                   int64       `json:"id"`
	Name                 string      `json:"name"`
	CourseId             int64       `json:"courseId"`
	CourseName           string      `json:"courseName"`
	ObligatoryType       int         `json:"obligatoryType"`
	Price                float64     `json:"price"`
	CompositeTermId      int64       `json:"compositeTermId"`
	StartTime            int64       `json:"startTime"`
	EndTime              int64       `json:"endTime"`
	Position             int         `json:"position"`
	TermType             int         `json:"termType"`
	CoursePhoto          string      `json:"coursePhoto"`
	Outline              interface{} `json:"outline"`
	CourseIntroduction   string      `json:"courseIntroduction"`
	TermOnDemandVo       interface{} `json:"termOnDemandVo"`
	SingleTermScheduleVo struct {
		Id                   int64       `json:"id"`
		GmtCreate            int64       `json:"gmtCreate"`
		GmtModified          int64       `json:"gmtModified"`
		StartTime            int64       `json:"startTime"`
		EndTime              int64       `json:"endTime"`
		EnrollStartTime      interface{} `json:"enrollStartTime"`
		EnrollEndTime        interface{} `json:"enrollEndTime"`
		EnrollEndTimeType    int         `json:"enrollEndTimeType"`
		Duration             interface{} `json:"duration"`
		CourseLoad           interface{} `json:"courseLoad"`
		CloseVisableStatus   int         `json:"closeVisableStatus"`
		CloseVisableTime     int         `json:"closeVisableTime"`
		ExpectLessonDuration interface{} `json:"expectLessonDuration"`
	} `json:"singleTermScheduleVo"`
	AssignType      interface{} `json:"assignType"`
	CompositeType   interface{} `json:"compositeType"`
	PaperTextbook   interface{} `json:"paperTextbook"`
	ProviderId      interface{} `json:"providerId"`
	PublishStatus   interface{} `json:"publishStatus"`
	OnlineFlag      interface{} `json:"onlineFlag"`
	LearnScheduleVo struct {
		TermId             interface{} `json:"termId"`
		LessonTotalCount   int         `json:"lessonTotalCount"`
		UnitTotalCount     int         `json:"unitTotalCount"`
		HasLearnCount      int         `json:"hasLearnCount"`
		StartTime          int64       `json:"startTime"`
		EndTime            int64       `json:"endTime"`
		LessonPosition     interface{} `json:"lessonPosition"`
		LastLearnUnitId    int64       `json:"lastLearnUnitId"`
		LastLearnUnitName  string      `json:"lastLearnUnitName"`
		LastLearnUnitType  int         `json:"lastLearnUnitType"`
		LastCourseId       interface{} `json:"lastCourseId"`
		LastCourseName     interface{} `json:"lastCourseName"`
		LastCompositeType  interface{} `json:"lastCompositeType"`
		LastLearnTermId    interface{} `json:"lastLearnTermId"`
		LastLearnTime      int64       `json:"lastLearnTime"`
		HasLearned         bool        `json:"hasLearned"`
		TermLearnProgress  float64     `json:"termLearnProgress"`
		GraduationStatus   int         `json:"graduationStatus"`
		HasStarted         int         `json:"hasStarted"`
		AllOptionsDisabled bool        `json:"allOptionsDisabled"`
		LessonName         interface{} `json:"lessonName"`
	} `json:"learnScheduleVo"`
	TxtOutline string `json:"txtOutline"`
	Options    int    `json:"options"`

	chapters []*Chapter
}

type CataLogList struct {
	CataLogList []*Chapter `json:"catalogList"`
}

type Chapter struct {
	Data struct {
		ReleaseTime interface{} `json:"releaseTime"`
		Description string      `json:"description"`
	} `json:"data"`
	Children    []*Child    `json:"children"`
	DraftStatus interface{} `json:"draftStatus"`
	Name        string      `json:"name"`
	Id          int64       `json:"id"`
	Type        int         `json:"type"`
	QuizLists   interface{} `json:"quizLists"`
}

type Data struct {
	ViewPriviledge   int         `json:"viewPriviledge"`
	LastLearnTime    interface{} `json:"lastLearnTime"`
	ContentId        int         `json:"contentId"`
	PaperData        interface{} `json:"paperData"`
	LiveStartTime    interface{} `json:"liveStartTime"`
	LiveFinishedTime interface{} `json:"liveFinishedTime"`
	ViewStatus       int         `json:"viewStatus"`
	ContentStatus    interface{} `json:"contentStatus"`
	StudyType        interface{} `json:"studyType"`
	ClientVisible    int         `json:"clientVisible"`
	JsonContent      interface{} `json:"jsonContent"`
	ContentType      int         `json:"contentType"`
	TargetUrl        interface{} `json:"targetUrl"`
	LiveStatus       interface{} `json:"liveStatus"`
}

type Child struct {
	Data        Data        `json:"data"`
	Children    []*Child    `json:"children"`
	DraftStatus interface{} `json:"draftStatus"`
	Name        string      `json:"name"`
	Id          int64       `json:"id"`
	Type        int         `json:"type"`
}
