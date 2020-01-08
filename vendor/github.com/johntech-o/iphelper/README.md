# iphelper
A ip-database library  for golang. Find a information of an IP address including country,province,city,region,street and network provider.

u can find example in the iphelper_test.go 


// init use the ip.dat file

	var store = NewIpStore("ip.dat")


// get geo info of ip address

	geo, e := store.GetGeoByIp("")

//  map[country:中国 province:北京市 city:北京市 zone:朝阳区 location:未知 operator:电信 areacode:20003000100370101]

	
	fmt.Println(geo)


//  get areacode of ip address

	code, e := store.GetGeocodeByIp("")

	// 20003000100370101

	fmt.Println(code)

// get the geo info of areacode

// u can save the areacode to user`s session

// get the geo info by areacode is more fast than by ip address

	codeGeo := store.GetGeoByGeocode(code)

//  map[country:中国 province:上海市 city:上海市 zone:未知 location:未知 operator:未知 areacode:20017009000000100]

	fmt.Println(codeGeo)


// get all the area info of the ip databases
	
	table := store.GetMetaTable()
	
	for typ, areas := range table {
		fmt.Println(typ, len(areas), areas)
		fmt.Println("--------------------")
	}
