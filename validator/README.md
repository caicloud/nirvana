# Validator

## About

We use [go-playground/validator.v9](https://github.com/go-playground/validator/tree/v9)
which provides flexible and rich methods to validate structs and values for validation.

## Design

Validator makes use of `Operators` in `definition.Parameter` struct.

`type Operator func(ctx context.Context, object interface{}) (interface{}, error)`

Each validator is a Operator. Nirvana provided 3 preset validators which using validator.v9 inside,
so you can use [validators and tags provided by go-playground/validator.v9](https://godoc.org/gopkg.in/go-playground/validator.v9#hdr-Baked_In_Validators_and_Tags) directly.

`validator.Struct` for struct value validation.

`validator.Var` and `validator.VarWithValue` for single variable validation.

## Custom Validator

More complicated validation can be achieved by writing your own validator like this:

```
import "github.com/caicloud/nirvana/validator"

type Booking struct {
	UserName string    `json:"name" validate:"required,printascii"`
	CheckIn  time.Time `json:"checkIn" validate:"required"`
	CheckOut time.Time `json:"checkOut" validate:"required,gtfield=CheckIn"`
}

func validateBooking(ctx context.Context, obj interface{}) (interface{}, error) {
	// Note we can still use validator pkg to check before our custom validation code.
	obj, err := validator.Struct()(ctx, obj)
	if err != nil {
		return obj, err
	}

	// for some cases that can not be coverd by basic tags in validator,
	// we can write validation code here.
	b := obj.(*Booking)
	today := time.Now()
	if today.Year() > b.CheckIn.Year() || today.YearDay() > b.CheckIn.YearDay() {
		return obj, fmt.Errorf("checkIn %s is not valid", b.CheckIn.String())
	}
	return obj, nil
}

var desc = definition.Descriptor{
	Path:        "/api/v1/",
	Definitions: []definition.Definition{},
	Consumes:    []string{"application/json"},
	Produces:    []string{"application/json"},
	Children: []definition.Descriptor{
		{
			Path: "/booking",
			Definitions: []definition.Definition{
				{
					Method:   definition.Create,
					Function: Handle,
					Parameters: []definition.Parameter{
						{
							Source:    definition.Body,
							Name:      "booking",
							Operators: []definition.Operator{validateBooking}, // use validateBooking to validate body
						},
					},
					Results: []definition.Result{
						{Type: definition.Data},
						{Type: definition.Error},
					},
				},
			},
		},
	},
}

```
