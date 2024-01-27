/*
objectstorage パッケージは、オブジェクトストレージを扱うためのパッケージです。
*/

package objectstorage

//TODO: ExistsObjectの挙動を実験するときに使用した、仮のテスト

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
