Spanner Schema

```
EventId	INT64
PlayerId	STRING(120)
Timestamp	INT64
EventType	STRING(120)
Data	STRING(120)
LastUpdated	TIMESTAMP
```

Example data structure

```
EventId: 1676994783 // time.Now().Unix()
PlayerId: "abc_1" // where 1 is cid (1 or 2)
Timestamp: 1676994783 // time.Now().Unix()
EventType: "SpawnMissile" // SpawnMissile of DestroyEvent
Data: nid:5577006791947779412 owner:5577006791947779410 pos:<x:-3.7588067 y:-10.875368 > momentum:<x:3.8375332 y:11.067178 > rot:1.5707964 // dictionary
```

To populate Spanner

gcloud spanner databases execute-sql spaceagon-db-demo --instance=spaceagon-demo \
  --sql="INSERT gameevents (EventId, PlayerID, Timestamp, EventType, Data, LastUpdated)
  VALUES (1, '2_jfb', 1676994786, 'SpawnMissile', 'none', CURRENT_TIMESTAMP())"


gcloud spanner databases execute-sql spaceagon-db-demo --instance=spaceagon-demo \
  --sql="INSERT gameevents (EventId, PlayerID, Timestamp, EventType, Data, LastUpdated)
  VALUES (2, '2_jfb', 1676994784, 'SpawnMissile', 'none', CURRENT_TIMESTAMP())"

gcloud spanner databases execute-sql spaceagon-db-demo --instance=spaceagon-demo \
  --sql="INSERT gameevents (EventId, PlayerID, Timestamp, EventType, Data, LastUpdated)
  VALUES (3, '2_jfb', 1676994791, 'SpawnMissile', 'none', CURRENT_TIMESTAMP())"

gcloud spanner databases execute-sql spaceagon-db-demo --instance=spaceagon-demo \
  --sql="INSERT gameevents (EventId, PlayerID, Timestamp, EventType, Data, LastUpdated)
  VALUES (4, '1_meb', 1676994799, 'SpawnMissile', 'none', CURRENT_TIMESTAMP())"


Example results
{"data":[{"Name":"1_meb","Score":1},{"Name":"2_jfb","Score":2}]}/
