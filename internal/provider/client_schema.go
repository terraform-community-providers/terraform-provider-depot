package provider

type Project struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

type ProjectOutput struct {
	Project Project `json:"project"`
}

type ProjectCreateInputProject struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type ProjectCreateInput struct {
	Project ProjectCreateInputProject `json:"project"`
}

type ProjectCreateOutput struct {
	Project Project `json:"project"`
}

type ProjectUpdateInputProject struct {
	Name string `json:"name"`
}

type ProjectUpdateInput struct {
	Project ProjectUpdateInputProject `json:"project"`
}
