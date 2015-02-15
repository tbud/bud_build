package script

import (
	"text/template"
)

func processScript() {
	temp, err := template.ParseFiles("")
	panicOnError(err, "Failed to parse template %s", srcFile)

	dst, err := os.Create(destFile)
	panicOnError(err, "Failed to create file %s", dst.Name())

	panicOnError(temp.Execute(dst, data), "Failed to render template %s", srcFile)

	panicOnError(dst.Close(), "Failed to close file %s", dst.Name())

}
