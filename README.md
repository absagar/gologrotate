# gologrotate
Rotate logs generated by your golang app and upload them to cloud.

Works similar to the logrotate utility. Avoids adding rules for logrotate for every machine you deploy your app. Can also upload to the cloud (includes an example of gce adapter) after they have been compressed.

The "Classic()" function initializes the package with some sane defaults. Keep the CopyTruncate parameter as true if you are not sure what that means.

How to use :

		lrc := logrotate.Classic([]string{"logfile1", "logfile2"})
		lrc.Init()

If cloud upload functionality is also needed, either write an adapter for your provider or use the gce adapter provided like this :

		lrcg := &logrotate.GceConfig{
			CredentialsFile: "path_to_your_credentials_file",
			Bucket:          "bucket_name",
			Location:        "location",
			Scope:           "https://www.googleapis.com/auth/devstorage.full_control",
		}
		lrc.BackupConfig = &logrotate.BackupConfig{
			BackupFunc: lrcg.Backup,
			MaxBackups: 2,
		}

