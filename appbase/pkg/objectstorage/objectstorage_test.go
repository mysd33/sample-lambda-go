/*
objectstorage パッケージは、オブジェクトストレージを扱うためのパッケージです。
*/

package objectstorage

// TODO: S3アクセスの挙動を実験するときに使用した仮のテストなのでコメントアウト
/*
func Test_defaultObjectStorageAccessor_List(t *testing.T) {
	//bucketName := "samplebucket123"
	bucketName := "mysd33bucket123"
	myCfg := config.NewTestConfig(map[string]string{
		//	"S3_LOCAL_ENDPOINT": "http://host.docker.internal:9000",
	})
	messageSource, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(messageSource)
	objectStorageAccessor, _ := NewObjectStorageAccessor(
		myCfg,
		logger,
	)

	type args struct {
		bucketName string
		folderPath string
	}
	tests := []struct {
		name string
		a    ObjectStorageAccessor
		args args
		// TODO assert
		//want    []types.Object
		wantErr bool
	}{
		// Add test cases.
		{"test1", objectStorageAccessor, args{bucketName, "input"}, false},
	}
	for _, tt := range tests {
		//  テスト用にX-Rayのセグメント開始
		ctx, seg := xray.BeginSegment(context.Background(), "objectstorage_test")
		apcontext.Context = ctx
		defer seg.Close(nil)
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.List(tt.args.bucketName, tt.args.folderPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultObjectStorageAccessor.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, v := range got {
				fmt.Println(*v.Key)
			}
			// TODO assert
			//if !reflect.DeepEqual(got, tt.want) {
			//		t.Errorf("defaultObjectStorageAccessor.List() = %v, want %v", got, tt.want)
			//}
		})
	}
}
*/

/*
func Test_defaultObjectStorageAccessor_ExistsObject(t *testing.T) {
	myCfg := config.NewTestConfig(map[string]string{
		"S3_LOCAL_ENDPOINT": "http://host.docker.internal:9000",
	})
	messageSource, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(messageSource, myCfg)
	objectStorageAccessor, _ := NewObjectStorageAccessor(
		myCfg,
		logger,
	)

	type args struct {
		bucketName string
		objectKey  string
	}
	tests := []struct {
		name    string
		a       ObjectStorageAccessor
		args    args
		want    bool
		wantErr bool
	}{
		// Add test cases.
		{"test1", objectStorageAccessor, args{"samplebucket123", "test"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//  テスト用にX-Rayのセグメント開始
			ctx, seg := xray.BeginSegment(context.Background(), "objectstorage_test")
			apcontext.Context = ctx
			defer seg.Close(nil)

			got, err := tt.a.ExistsObject(tt.args.bucketName, tt.args.objectKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultObjectStorageAccessor.ExistsObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("defaultObjectStorageAccessor.ExistsObject() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/

/*
func Test_defaultObjectStorageAccessor_UploadFromReader(t *testing.T) {
	myCfg := config.NewTestConfig(map[string]string{
		//"S3_LOCAL_ENDPOINT": "http://host.docker.internal:9000",
	})
	messageSource, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(messageSource, myCfg)
	objectStorageAccessor, _ := NewObjectStorageAccessor(
		myCfg,
		logger,
	)

	//file, _ := os.Open("C:\\tmp\\todolist.csv")
	file, _ := os.Open("C:\\tmp\\IMG_4438.JPG")
	defer file.Close()

	type args struct {
		bucketName string
		objectKey  string
		reader     io.Reader
	}
	tests := []struct {
		name    string
		a       ObjectStorageAccessor
		args    args
		wantErr bool
	}{
		// Add test cases.
		//{"test1", objectStorageAccessor, args{"samplebucket123", "test.csv", file}, false},
		//{"test2", objectStorageAccessor, args{"samplebucket123", "test.jpg", file}, false},
		//{"test1", objectStorageAccessor, args{"mysd33bucket123", "test.csv", file}, false},
		{"test2", objectStorageAccessor, args{"mysd33bucket123", "test.jpg", file}, false},
	}
	for _, tt := range tests {
		//  テスト用にX-Rayのセグメント開始
		ctx, seg := xray.BeginSegment(context.Background(), "objectstorage_test")
		apcontext.Context = ctx
		defer seg.Close(nil)

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.UploadFromReader(tt.args.bucketName, tt.args.objectKey, tt.args.reader); (err != nil) != tt.wantErr {
				t.Errorf("defaultObjectStorageAccessor.UploadLargeObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}*/

/*
func Test_defaultObjectStorageAccessor_Copy(t *testing.T) {
	//bucketName := "samplebucket123"
	bucketName := "mysd33bucket123"

	myCfg := config.NewTestConfig(map[string]string{
		//	"S3_LOCAL_ENDPOINT": "http://host.docker.internal:9000",
	})

	messageSource, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(messageSource)
	objectStorageAccessor, _ := NewObjectStorageAccessor(
		myCfg,
		logger,
	)
	type args struct {
		bucketName       string
		objectKey        string
		targetFolderPath string
	}
	tests := []struct {
		name    string
		a       ObjectStorageAccessor
		args    args
		wantErr bool
	}{
		// Add test cases.
		{"test1", objectStorageAccessor, args{bucketName, "input/todolist.csv", "output"}, false},
		{"test2", objectStorageAccessor, args{bucketName, "input/ほげほげ.csv", "output"}, false},
	}
	for _, tt := range tests {
		//  テスト用にX-Rayのセグメント開始
		ctx, seg := xray.BeginSegment(context.Background(), "objectstorage_test")
		apcontext.Context = ctx
		defer seg.Close(nil)

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.Copy(tt.args.bucketName, tt.args.objectKey, tt.args.targetFolderPath); (err != nil) != tt.wantErr {
				t.Errorf("defaultObjectStorageAccessor.Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/

/*
func Test_defaultObjectStorageAccessor_CopyFolder(t *testing.T) {
	bucketName := "samplebucket123"
	//bucketName := "mysd33bucket123"
	myCfg := config.NewTestConfig(map[string]string{
		"S3_LOCAL_ENDPOINT": "http://host.docker.internal:9000",
	})
	messageSource, _ := message.NewMessageSource()
	logger, _ := logging.NewLogger(messageSource)
	objectStorageAccessor, _ := NewObjectStorageAccessor(
		myCfg,
		logger,
	)
	type args struct {
		bucketName       string
		srcFolderPath    string
		targetFolderPath string
		nested           bool
	}
	tests := []struct {
		name    string
		a       ObjectStorageAccessor
		args    args
		wantErr bool
	}{
		// Add test cases.
		{"test1", objectStorageAccessor, args{bucketName, "input", "output", true}, false},
	}
	for _, tt := range tests {
		//  テスト用にX-Rayのセグメント開始
		ctx, seg := xray.BeginSegment(context.Background(), "objectstorage_test")
		apcontext.Context = ctx
		defer seg.Close(nil)

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.CopyFolder(tt.args.bucketName, tt.args.srcFolderPath, tt.args.targetFolderPath, tt.args.nested); (err != nil) != tt.wantErr {
				t.Errorf("defaultObjectStorageAccessor.CopyFolder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
