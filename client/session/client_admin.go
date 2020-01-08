package session

/*
 * get session qps from specified saver
 * @param gorpcAddr is specified saver go rpc address
 * @return (*SessionStat, nil) if successful, otherwise (*SessionStat{}, error) will be return
 */
func GetSessionQps(gorpcAddr string) (*SessionStat, error) {
	resp := &SessionStat{}
	if err := GorpcClient.CallWithAddress(gorpcAddr, "GorpcService", "GetSessionQps", 0, resp); err != nil {
		return &SessionStat{}, err
	}

	return resp, nil
}

/*
 * get session total operations count after start from specified saver
 * @param gorpcAddr is specified saver go rpc address
 * @return (*SessionStat, nil) if successful, otherwise (*SessionStat{}, error) will be return
 */
func GetSessionTotalOps(gorpcAddr string) (*SessionStat, error) {
	resp := &SessionStat{}
	if err := GorpcClient.CallWithAddress(gorpcAddr, "GorpcService", "GetSessionTotalOps", 0, resp); err != nil {
		return &SessionStat{}, err
	}

	return resp, nil
}
