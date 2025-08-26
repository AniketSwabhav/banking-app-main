package docs

import "github.com/swaggo/swag"

const docTemplate = `{
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "2611",
	BasePath:         "/api/v1/contact-app",
	Schemes:          []string{},
	Title:            "contact-app",
	Description:      "Short Description",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
