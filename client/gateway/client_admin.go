package gateway

/*
 * get gateway qps from specified saver
 * @param gatewayGorpcAddr is specified saver go rpc address
 * @return (*GatewayStat, nil) if successful, otherwise (*GatewayStat{}, error) will be return
 */
func GetGatewayQps(gatewayGorpcAddr string) (*GatewayStat, error) {
	resp := &GatewayStat{}
	if err := GorpcClient.CallWithAddress(gatewayGorpcAddr, "GorpcService", "GetGatewayQps", 0, resp); err != nil {
		return &GatewayStat{}, err
	}

	return resp, nil
}

/*
 * get gateway total operations count after start from specified saver
 * @param gatewayGorpcAddr is specified saver go rpc address
 * @return (*GatewayStat, nil) if successful, otherwise (*GatewayStat{}, error) will be return
 */
func GetGatewayTotalOps(gatewayGorpcAddr string) (*GatewayStat, error) {
	resp := &GatewayStat{}
	if err := GorpcClient.CallWithAddress(gatewayGorpcAddr, "GorpcService", "GetGatewayTotalOps", 0, resp); err != nil {
		return &GatewayStat{}, err
	}

	return resp, nil
}
