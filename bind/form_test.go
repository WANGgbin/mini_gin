package bind

import (
	"github.com/smartystreets/goconvey/convey"
	"net/url"
	"testing"
)

func Test_bindUrlValuesToStruct(t *testing.T) {
	convey.Convey("", t, func(){
		type Person struct {
			private string
			Name string `form:"name"`
			Age int `form:"age,default=100"`
			Weight float64
			Hobbies []string `form:"hobbies"`
			Scores []int `form:"scores,default=100"`
		}

		testCases := []struct{
			vals url.Values
			wantErr bool
		}{
			{
				vals: map[string][]string{
					"name": {"xiaoming"},
					"age": {"10"},
					"Weight": {"66.6"},
					"hobbies": {"sleep", "eat"},
					"scores": {"0", "1"},
				},
			},
			{
				vals: map[string][]string{
					"name": {"xiaoming"},
					"age": {"10"},
					"Weight": {"66.6"},
					"hobbies": {"sleep", "eat"},
					"scores": {"str", "1"},
				},
				wantErr: true,
			},
			{
				vals: map[string][]string{
					"name": {"xiaoming"},
					"Weight": {"66.6"},
					"hobbies": {"sleep", "eat"},
				},
			},
		}

		for _, testCase := range testCases {
			var p Person
			gotErr := bindUrlValuesToStruct(testCase.vals, &p)
			if testCase.wantErr {
				convey.So(gotErr, convey.ShouldNotBeNil)
				t.Log(gotErr)
			} else {
				convey.So(gotErr, convey.ShouldBeNil)
				t.Logf("%+v", p)
			}
		}
	})
}

