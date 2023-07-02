package mini_gin

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_validateRoute(t *testing.T) {
	convey.Convey("", t, func(){
		testCases := []struct{
			route string
			wantResult bool
		}{
			{
				route: "/a/b/c",
				wantResult: true,
			},
			{
				route: "/:key",
				wantResult: true,
			},
			{
				route: "/:key/",
				wantResult: true,
			},
			{
				route: "/:/",
				wantResult: false,
			},
			{
				route: "//",
				wantResult: false,
			},
			{
				route: "a/b",
				wantResult: false,
			},
			{
				route: "",
				wantResult: false,
			},
			{
				route: "/:ke:y/",
				wantResult: false,
			},
			{
				route: "/prefix:key",
				wantResult: true,
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.route, func(){
				gotResult := validateRoute(testCase.route)
				convey.So(gotResult, convey.ShouldEqual, testCase.wantResult)
			})

		}
	})
}

func Test_validateSegment(t *testing.T) {
	convey.Convey("", t, func(){
		testCases := []struct{
			seg string
			valid bool
		}{
			{
				seg: "abc",
				valid: true,
			},
			{
				seg: "prefix:key",
				valid: true,
			},
			{
				seg: "",
				valid: false,
			},
			{
				seg: ":",
				valid: false,
			},
			{
				seg: "prefix:",
				valid: false,
			},
			{
				seg: ":key1:key2",
				valid: false,
			},
		}

		for _, testCase := range testCases {
			convey.Convey(testCase.seg, func(){
				gotResult := validateSegment(testCase.seg)
				convey.So(gotResult, convey.ShouldEqual, testCase.valid)
			})
		}
	})
}