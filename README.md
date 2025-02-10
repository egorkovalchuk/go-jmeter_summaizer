# JMeter Summariser to InfluxDB Metrics Exporter
This project is designed to transfer metrics from JMeter Summariser files (*.out) to InfluxDB. The program reads the files, extracts metrics, and sends them to InfluxDB in a format suitable for analysis and visualization.

# Key Features
Reading ***.out** files created by JMeter Summariser.

Extracting metrics such as:

* Average response time (Avg).
* Minimum response time (Min).
* Maximum response time (Max).
* Number of errors (Err).
* Error percentage (Err%).
* Active threads (Active).
* Started threads (Started).
* Finished threads (Finished).

Sending metrics to InfluxDB in two modes:
* **Delta** — for intermediate results.
* **Total** — for final results.

# Example of Operation
## Input Data (file sso_auth-1233414.out)

```plaintext
summary +    214 in 00:00:30 =    7.1/s Avg:    18 Min:     0 Max:   593 Err:   105 (49.07%) Active: 17 Started: 17 Finished: 0
summary = 224168 in 08:43:40 =    7.1/s Avg:   104 Min:     0 Max: 519940 Err: 115641 (51.59%)
```

## Output Data (InfluxDB)
**Delta** (intermediate results)
```plaintext
delta,project=MyProject,suite=sso_auth avg=18,min=0,max=593,rate=7.1,err=105,errpct=49.07,ath=17,sth=17,eth=0
```
**Total** (final results)
```plaintext
total,project=MyProject,suite=sso_auth avg=104,min=0,max=519940,rate=7.1,err=115641,errpct=51.59,ath=0,sth=0,eth=0
```

# Launch Parameters
-debug — enables debug mode. Outputs additional information to the console.

-hp — uses the HP package. Works correctly on Windows and Linux. If not used, the program works only on Linux (file locking occurs on Windows).

-noinf — disables writing to InfluxDB. Used for testing purposes.

# Example of InfluxDB Configuration
Make sure a bucket is created in InfluxDB and access permissions are configured. Example query to create a bucket:

```sql
CREATE BUCKET "jmeter_metrics" WITH RETENTION POLICY "30d"
```

# Workflow Logic
1. The program reads ***.out** files line by line.
2. For each line starting with **summary +** or **summary =**, metrics are extracted.
3. Metrics are transformed into a format understandable by InfluxDB.
4. Data is sent to InfluxDB using the Line Protocol.

# Alternative Option
Use Telegraf to read parameters. Example in **telegraf.conf**:
(You can provide an example configuration for Telegraf here if needed.)