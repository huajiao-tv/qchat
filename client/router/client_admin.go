package router

/*
 * get router qps from specified saver
 * @param gorpcAddr is specified saver go rpc address
 * @return (*RouterStat, nil) if successful, otherwise (*RouterStat{}, error) will be return
 */
func GetRouterQps(gorpcAddr string) (*RouterStat, error) {
	resp := &RouterStat{}
	if err := GorpcClient.CallWithAddress(gorpcAddr, "GorpcService", "GetRouterQps", 0, resp); err != nil {
		return &RouterStat{}, err
	}

	return resp, nil
}

/*
 * get router total operations count after start from specified saver
 * @param gorpcAddr is specified saver go rpc address
 * @return (*RouterStat, nil) if successful, otherwise (*RouterStat{}, error) will be return
 */
func GetRouterTotalOps(gorpcAddr string) (*RouterStat, error) {
	resp := &RouterStat{}
	if err := GorpcClient.CallWithAddress(gorpcAddr, "GorpcService", "GetRouterTotalOps", 0, resp); err != nil {
		return &RouterStat{}, err
	}

	return resp, nil
}
