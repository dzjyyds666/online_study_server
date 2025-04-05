package core

var UserRole = struct {
	Admin   int
	Teacher int
	Student int
}{
	Admin:   3,
	Teacher: 2,
	Student: 1,
}

var UserStatus = struct {
	Active   int8
	Inactive int8
	Deleted  int8
}{
	Active:   0,
	Inactive: 1,
	Deleted:  2,
}
