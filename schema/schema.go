package main

import (
	"github.com/iancoleman/strcase"
	"github.com/swaggest/jsonschema-go"
	"google.golang.org/protobuf/proto"
)

/*
 * propertyToCamelInterceptor
 */
func propertyToCamelInterceptor(params jsonschema.InterceptSchemaParams) (stop bool, err error) {

	if len(params.Schema.Properties) > 0 {
		updateProps(params)
	}

	updateRequired(params)

	return false, nil
}

/*
 * updateProps
 */
func updateProps(params jsonschema.InterceptSchemaParams) {

	props := map[string]jsonschema.SchemaOrBool{}
	for k, p := range params.Schema.Properties {
		props[strcase.ToLowerCamel(k)] = p
	}

	params.Schema.Properties = props
	params.Schema.AdditionalProperties = &jsonschema.SchemaOrBool{TypeBoolean: proto.Bool(false)}
}

/*
 * updateRequired
 */
func updateRequired(params jsonschema.InterceptSchemaParams) {

	required := []string{}
	for _, s := range params.Schema.Required {
		required = append(required, strcase.ToLowerCamel(s))
	}

	params.Schema.Required = required
}

/*
 * GetSchema
 */
func GetSchema(i any) (jsonschema.Schema, error) {

	reflector := jsonschema.Reflector{}

	return reflector.Reflect(i,
		jsonschema.InlineRefs,
		jsonschema.ProcessWithoutTags,
		jsonschema.InterceptSchema(propertyToCamelInterceptor),
	)
}
