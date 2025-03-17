package core

type StudyClass struct {
	SCid          string `json:"sc_id"`
	ClassName     string `json:"class_name"`
	StudentNumber int64  `json:"student_number"`
	Cid           string `json:"cid"`
}

func (s *StudyClass) WithSCid(id string) *StudyClass {
	s.SCid = id
	return s
}

func (s *StudyClass) WithClassName(name string) *StudyClass {
	s.ClassName = name
	return s
}

func (s *StudyClass) WithStudentNumber(number int64) *StudyClass {
	s.StudentNumber = number
	return s
}
