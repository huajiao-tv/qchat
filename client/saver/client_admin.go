package saver

/*
 * get saver qps from specified saver
 * @param saverGorpcAddr is specified saver go rpc address
 * @return (*SaverStat, nil) if successful, otherwise (*SaverStat{}, error) will be return
 */
func GetSaverQps(saverGorpcAddr string) (*SaverStat, error) {
	resp := &SaverStat{}
	if err := GorpcClient.CallWithAddress(saverGorpcAddr, "GorpcService", "GetSaverQps", 0, resp); err != nil {
		return &SaverStat{}, err
	}

	return resp, nil
}

/*
 * get saver total operations count after start from specified saver
 * @param saverGorpcAddr is specified saver go rpc address
 * @return (*SaverStat, nil) if successful, otherwise (*SaverStat{}, error) will be return
 */
func GetSaverTotalOps(saverGorpcAddr string) (*SaverStat, error) {
	resp := &SaverStat{}
	if err := GorpcClient.CallWithAddress(saverGorpcAddr, "GorpcService", "GetSaverTotalOps", 0, resp); err != nil {
		return &SaverStat{}, err
	}

	return resp, nil
}
