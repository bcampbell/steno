// +build !darwin
package kludge

func DataPath() (string, error) {
	return ".", nil
}
