// Package errors is designed to be the default error returned by nirvana.
// Recommended usage example:
//
// import "github.com/caicloud/nirvana/errors"

// type App struct {}

// var AppIsNotFound = errors.NotFound.NewFactory(
// 	errors.Reason("app-admin:app-not-found"),
// 	"server can't find app ${appName} in partition ${partitionName}",
// )

// func GetApp() (*App, error) {
// 	return nil, AppIsNotFound.New("mongodb", "test")
// }

// func main() {
// 	_, err := GetApp()
// 	if AppIsNotFound.CanNew(err) {
// 		// Do something
// 	}
// }

package errors
