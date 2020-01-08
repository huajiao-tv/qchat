package center

/*
 * get center qps from specified saver
 * @param centerGorpcAddr is specified saver go rpc address
 * @return (*CenterStat, nil) if successful, otherwise (*CenterStat{}, error) will be return
 */
func GetCenterQps(centerGorpcAddr string) (*CenterStat, error) {
	resp := &CenterStat{}
	if err := GorpcClient.CallWithAddress(centerGorpcAddr, "GorpcService", "GetCenterQps", 0, resp); err != nil {
		return &CenterStat{}, err
	}

	return resp, nil
}

/*
 * get center total operations count after start from specified saver
 * @param centerGorpcAddr is specified saver go rpc address
 * @return (*CenterStat, nil) if successful, otherwise (*CenterStat{}, error) will be return
 */
func GetCenterTotalOps(centerGorpcAddr string) (*CenterStat, error) {
	resp := &CenterStat{}
	if err := GorpcClient.CallWithAddress(centerGorpcAddr, "GorpcService", "GetCenterTotalOps", 0, resp); err != nil {
		return &CenterStat{}, err
	}

	return resp, nil
}
