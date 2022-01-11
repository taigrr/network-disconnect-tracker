To set partition expiration, run the following sql statement:
```
ALTER SCHEMA `project_name.DatasetName`
 SET OPTIONS(
     default_partition_expiration_days=365
 )
```
