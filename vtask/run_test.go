// @File     : run_test
// @Author   : Ville
// @Time     : 19-10-9 上午9:50
// vtask
package vtask

//func TestRun(t *testing.T) {
//	Run()
//
//	Add(1234, func(params interface{}) error {
//		fmt.Println(params)
//		return nil
//	})
//
//	Add("hello is me", func(params interface{}) error {
//		fmt.Println(params)
//		return nil
//	})
//
//	Add(map[string]string{"one": "过来", "two": "@#￥"}, func(params interface{}) error {
//		fmt.Println(params)
//		return errors.New("aaa")
//	})
//
//	Add(map[string]string{"one": "出错重试一次", "two": "@#￥"}, func(params interface{}) error {
//		fmt.Println(params)
//		return errors.New("aaa")
//	})
//
//	go func() {
//		for i := 0; i < 100; i++ {
//			Add(i, func(params interface{}) error {
//				fmt.Println(params)
//				return nil
//			})
//		}
//	}()
//
//	for !CanStop() {
//		time.Sleep(time.Millisecond * 2)
//	}
//
//	Stop()
//}
