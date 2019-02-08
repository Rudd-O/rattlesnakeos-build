package stack

func RenderTemplate(templateStr string, params interface{}) ([]byte, error) {
	return renderTemplate(templateStr, params)
}
