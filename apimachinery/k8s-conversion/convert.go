package main

import (
	convert "cloudcome.net/k8s-conversion/funs"
)


func main() {

	//Conversion funs
	//convert.TestConverter_byteSlice()
	//convert.TestConverter_CallsRegisteredFunctions()
	//convert.TestConverter_DefaultConvert()
	//convert.TestConverter_IgnoredConversion()
	convert.TestConverter_meta()

}
