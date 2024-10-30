Project for transferring metrics from jmeter summariser to influxdb

Read *.out files

Example 

summary +    214 in 00:00:30 =    7.1/s Avg:    18 Min:     0 Max:   593 Err:   105 (49.07%) Active: 17 Started: 17 Finished: 0

to

delta,project=wf,suite=sso_wl_ccmp_auth avg=18,min=0,max=593,rate=7.1,err=105,errpct=49.07,ath=17,sth=17,eth=0

summary = 224168 in 08:43:40 =    7.1/s Avg:   104 Min:     0 Max: 519940 Err: 115641 (51.59%)

to

total,project=wf,suite=sso_wl_ccmp_auth avg=104,min=0,max=519940,rate=7.1,err=115641,errpct=51.59,ath=0,sth=0,eth=0 


Launch parameters

-debug Debug mode

-hp - Using the package from HP, works correctly in Win and Linux. If you do not use it, it works correctly only in Linux, in Win there is a file lock

-noinf disables writing to influx, for testing

