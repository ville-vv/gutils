package httpc

func ParseHttpParamForGet(inParam map[string]string) (outStr string) {
	if len(inParam) == 0 {
		return ""
	}
	isFirst := true
	tmpStr := ""
	for mk, mv := range inParam {
		tmpStr = mk + "=" + mv
		if isFirst {
			outStr = tmpStr
			isFirst = false
		}
		outStr = outStr + "&" + tmpStr
	}
	return
}
