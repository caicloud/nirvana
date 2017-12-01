// Package validator usage:

// type Application struct {
// 	Name      string `json:"name" validate:"required,printascii"`
// 	Namespace string `json:"namespace"`
// }

// Definitions: []definition.Definition{
// 	{
// 		Method:   definition.Create,
// 		Function: Handle,
// 		Parameters: []definition.Parameter{
// 			{
// 				Source:    definition.Query,
// 				Name:      "target1",
// 				Operators: []definition.Operator{validator.Var("gt=0,lt=10")},
// 			},
// 			{
// 				Source: definition.Body,
// 				Name:   "app",
// 				Operators: []definition.Operator{validator.Struct()},
// 			},
// 		},
// 	},
// },

package validator
