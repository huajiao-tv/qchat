package com.huajiao.comm.im.rpc;

public class RegisterServiceFilter extends Cmd {

	private static final long serialVersionUID = -8147965238409257718L;
	
	protected String _filter_classes;

	/**
	 * @param filter_classes
	 * @throws IllegalArgumentException
	 */
	public RegisterServiceFilter(String filter_classes) {
		super(Cmd.CMD_REGISTER_FILTER_SERVICE);
		if(filter_classes == null || filter_classes.length() == 0){
			throw new IllegalArgumentException();
		}
		this._filter_classes = filter_classes;
	}

	public String get_filter_classes() {
		return _filter_classes;
	}
	
}
