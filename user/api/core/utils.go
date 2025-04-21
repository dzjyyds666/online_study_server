package core

import (
	"github.com/dzjyyds666/opensource/logx"
	"github.com/xuri/excelize/v2"
	"io"
)

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

func parseExcel(sheetName string, r io.Reader) ([][]string, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		logx.GetLogger("study").Errorf("BetchAddStudentToClass|OpenReader Error|%v", err)
		return nil, err
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		logx.GetLogger("study").Errorf("BetchAddStudentToClass|GetRows Error|%v", err)
		return nil, err
	}
	return rows, nil
}
