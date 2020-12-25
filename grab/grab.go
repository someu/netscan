package grab

func AddModule(moduleName string, shortDescription string, longDescription string, port int, m ScanModule) (interface{}, error) {
	modules[moduleName] = m
	return nil, nil
}