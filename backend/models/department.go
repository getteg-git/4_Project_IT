package models

// ระดับการศึกษา (เช่น ปริญญาตรี, ปริญญาโท, ปริญญาเอก)
type DegreeLevel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ภาควิชา / สาขาวิชา
type Department struct {
	ID              int    `json:"id"`
	DegreeLevelID   int    `json:"degree_level_id"` // เชื่อมโยงกับระดับการศึกษา
	DegreeLevelName string `json:"degree_level_name,omitempty"`
	Name            string `json:"name"`
}
