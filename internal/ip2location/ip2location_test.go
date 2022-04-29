package ip2location

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pomerium/datasource/internal/jsonutil"
)

// Data comes from the free IP2Location LITE database. Attribution:
//   This site or product includes IP2Location LITE data available from
//   <a href="https://lite.ip2location.com">https://lite.ip2location.com</a>.

const sampleIP2LocationData = `
"0","281470681743359","-","-","-","-","0.000000","0.000000","-","-"
"281470681743360","281470698520575","-","-","-","-","0.000000","0.000000","-","-"
"281470698520576","281470698520831","US","United States of America","California","Los Angeles","34.052230","-118.243680","90001","-07:00"
"281470698520832","281470698521599","CN","China","Fujian","Fuzhou","26.061390","119.306110","350004","+08:00"
"281470698521600","281470698522623","AU","Australia","Victoria","Melbourne","-37.814000","144.963320","3000","+11:00"
"281470698522624","281470698524671","CN","China","Guangdong","Guangzhou","23.116670","113.250000","510140","+08:00"
"281470698524672","281470698528767","JP","Japan","Tokyo","Tokyo","35.689506","139.691700","160-0021","+09:00"
"281470698528768","281470698536959","CN","China","Guangdong","Guangzhou","23.116670","113.250000","510140","+08:00"
"281470698536960","281470698537983","JP","Japan","Hiroshima","Hiroshima","34.385280","132.455280","732-0057","+09:00"
"281470698537984","281470698538239","JP","Japan","Miyagi","Sendai","38.267000","140.867000","980-0802","+09:00"
"281470698538240","281470698541055","JP","Japan","Hiroshima","Hiroshima","34.385280","132.455280","732-0057","+09:00"
"281470698541056","281470698541311","JP","Japan","Shimane","Matsue","35.467000","133.050000","690-0015","+09:00"
"281470698541312","281470698541567","JP","Japan","Yamaguchi","Hikari","33.961940","131.942220","743-0021","+09:00"
"281470698541568","281470698541823","JP","Japan","Tottori","Yonago","35.433000","133.333000","683-0846","+09:00"
"281470698541824","281470698542079","JP","Japan","Tottori","Kurayoshi","35.433000","133.817000","682-0021","+09:00"
"281470698542080","281470698542335","JP","Japan","Tottori","Tottori","35.500000","134.233000","680-0805","+09:00"
"281470698542336","281470698542591","JP","Japan","Shimane","Matsue","35.467000","133.050000","690-0015","+09:00"
"281470698542592","281470698542847","JP","Japan","Okayama","Okayama","34.650000","133.917000","700-0824","+09:00"
"281470698542848","281470698543103","JP","Japan","Yamaguchi","Yamaguchi","34.183000","131.467000","754-0893","+09:00"
"281470698543104","281470698543359","JP","Japan","Shimane","Izumo","35.367000","132.767000","693-0044","+09:00"
"281470698543360","281470698543615","JP","Japan","Tottori","Kurayoshi","35.433000","133.817000","682-0021","+09:00"
"281470698543616","281470698543871","JP","Japan","Tottori","Tottori","35.500000","134.233000","680-0805","+09:00"
"281470698543872","281470698544127","JP","Japan","Shimane","Izumo","35.367000","132.767000","693-0044","+09:00"
"281470698544128","281470698544383","JP","Japan","Yamaguchi","Hikari","33.961940","131.942220","743-0021","+09:00"
"281470698544384","281470698544639","JP","Japan","Hiroshima","Hiroshima","34.385280","132.455280","732-0057","+09:00"
"281470698544640","281470698544895","JP","Japan","Tottori","Yonago","35.433000","133.333000","683-0846","+09:00"
"281470698544896","281470698545151","JP","Japan","Tottori","Kurayoshi","35.433000","133.817000","682-0021","+09:00"
"281470698545152","281470698545407","JP","Japan","Yamaguchi","Hikari","33.961940","131.942220","743-0021","+09:00"
"281470698545408","281470698545663","JP","Japan","Shimane","Izumo","35.367000","132.767000","693-0044","+09:00"
"281470698545664","281470698546175","JP","Japan","Yamaguchi","Yamaguchi","34.183000","131.467000","754-0893","+09:00"
"281470698546176","281470698546431","JP","Japan","Shimane","Izumo","35.367000","132.767000","693-0044","+09:00"
"281470698546432","281470698546687","JP","Japan","Shimane","Matsue","35.467000","133.050000","690-0015","+09:00"
"281470698546688","281470698547199","JP","Japan","Yamaguchi","Yamaguchi","34.183000","131.467000","754-0893","+09:00"
"281470698547200","281470698547455","JP","Japan","Yamaguchi","Hikari","33.961940","131.942220","743-0021","+09:00"
"281470698547456","281470698547711","JP","Japan","Shimane","Matsue","35.467000","133.050000","690-0015","+09:00"
"281470698547712","281470698547967","JP","Japan","Tottori","Tottori","35.500000","134.233000","680-0805","+09:00"
"281470698547968","281470698548223","JP","Japan","Yamaguchi","Yamaguchi","34.183000","131.467000","754-0893","+09:00"
"281470698548224","281470698548479","JP","Japan","Okayama","Okayama","34.650000","133.917000","700-0824","+09:00"
"281470698548480","281470698548735","JP","Japan","Hiroshima","Hiroshima","34.385280","132.455280","732-0057","+09:00"
"281470698548736","281470698548991","JP","Japan","Shimane","Izumo","35.367000","132.767000","693-0044","+09:00"
`

func TestParseCSV(t *testing.T) {
	var buf bytes.Buffer
	err := csvToJSON(jsonutil.NewJSONArrayStream(&buf), strings.NewReader(sampleIP2LocationData))
	require.NoError(t, err)
	assert.JSONEq(t, `[
  {
    "$index": {
      "cidr": "1.0.0.0/24"
    },
    "id": "1.0.0.0/24",
    "country": "US",
    "state": "California",
    "city": "Los Angeles",
    "zip": "90001",
    "timezone": "-07:00"
  },
  {
    "$index": {
      "cidr": "1.0.1.0/24"
    },
    "id": "1.0.1.0/24",
    "country": "CN",
    "state": "Fujian",
    "city": "Fuzhou",
    "zip": "350004",
    "timezone": "+08:00"
  },
  {
    "$index": {
      "cidr": "1.0.2.0/23"
    },
    "id": "1.0.2.0/23",
    "country": "CN",
    "state": "Fujian",
    "city": "Fuzhou",
    "zip": "350004",
    "timezone": "+08:00"
  },
  {
    "$index": {
      "cidr": "1.0.4.0/22"
    },
    "id": "1.0.4.0/22",
    "country": "AU",
    "state": "Victoria",
    "city": "Melbourne",
    "zip": "3000",
    "timezone": "+11:00"
  },
  {
    "$index": {
      "cidr": "1.0.8.0/21"
    },
    "id": "1.0.8.0/21",
    "country": "CN",
    "state": "Guangdong",
    "city": "Guangzhou",
    "zip": "510140",
    "timezone": "+08:00"
  },
  {
    "$index": {
      "cidr": "1.0.16.0/20"
    },
    "id": "1.0.16.0/20",
    "country": "JP",
    "state": "Tokyo",
    "city": "Tokyo",
    "zip": "160-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.32.0/19"
    },
    "id": "1.0.32.0/19",
    "country": "CN",
    "state": "Guangdong",
    "city": "Guangzhou",
    "zip": "510140",
    "timezone": "+08:00"
  },
  {
    "$index": {
      "cidr": "1.0.64.0/22"
    },
    "id": "1.0.64.0/22",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.68.0/24"
    },
    "id": "1.0.68.0/24",
    "country": "JP",
    "state": "Miyagi",
    "city": "Sendai",
    "zip": "980-0802",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.69.0/24"
    },
    "id": "1.0.69.0/24",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.70.0/23"
    },
    "id": "1.0.70.0/23",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.72.0/21"
    },
    "id": "1.0.72.0/21",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.80.0/24"
    },
    "id": "1.0.80.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Matsue",
    "zip": "690-0015",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.81.0/24"
    },
    "id": "1.0.81.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Hikari",
    "zip": "743-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.82.0/24"
    },
    "id": "1.0.82.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Yonago",
    "zip": "683-0846",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.83.0/24"
    },
    "id": "1.0.83.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Kurayoshi",
    "zip": "682-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.84.0/24"
    },
    "id": "1.0.84.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Tottori",
    "zip": "680-0805",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.85.0/24"
    },
    "id": "1.0.85.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Matsue",
    "zip": "690-0015",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.86.0/24"
    },
    "id": "1.0.86.0/24",
    "country": "JP",
    "state": "Okayama",
    "city": "Okayama",
    "zip": "700-0824",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.87.0/24"
    },
    "id": "1.0.87.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Yamaguchi",
    "zip": "754-0893",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.88.0/24"
    },
    "id": "1.0.88.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Izumo",
    "zip": "693-0044",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.89.0/24"
    },
    "id": "1.0.89.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Kurayoshi",
    "zip": "682-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.90.0/24"
    },
    "id": "1.0.90.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Tottori",
    "zip": "680-0805",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.91.0/24"
    },
    "id": "1.0.91.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Izumo",
    "zip": "693-0044",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.92.0/24"
    },
    "id": "1.0.92.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Hikari",
    "zip": "743-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.93.0/24"
    },
    "id": "1.0.93.0/24",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.94.0/24"
    },
    "id": "1.0.94.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Yonago",
    "zip": "683-0846",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.95.0/24"
    },
    "id": "1.0.95.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Kurayoshi",
    "zip": "682-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.96.0/24"
    },
    "id": "1.0.96.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Hikari",
    "zip": "743-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.97.0/24"
    },
    "id": "1.0.97.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Izumo",
    "zip": "693-0044",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.98.0/23"
    },
    "id": "1.0.98.0/23",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Yamaguchi",
    "zip": "754-0893",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.100.0/24"
    },
    "id": "1.0.100.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Izumo",
    "zip": "693-0044",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.101.0/24"
    },
    "id": "1.0.101.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Matsue",
    "zip": "690-0015",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.102.0/23"
    },
    "id": "1.0.102.0/23",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Yamaguchi",
    "zip": "754-0893",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.104.0/24"
    },
    "id": "1.0.104.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Hikari",
    "zip": "743-0021",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.105.0/24"
    },
    "id": "1.0.105.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Matsue",
    "zip": "690-0015",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.106.0/24"
    },
    "id": "1.0.106.0/24",
    "country": "JP",
    "state": "Tottori",
    "city": "Tottori",
    "zip": "680-0805",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.107.0/24"
    },
    "id": "1.0.107.0/24",
    "country": "JP",
    "state": "Yamaguchi",
    "city": "Yamaguchi",
    "zip": "754-0893",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.108.0/24"
    },
    "id": "1.0.108.0/24",
    "country": "JP",
    "state": "Okayama",
    "city": "Okayama",
    "zip": "700-0824",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.109.0/24"
    },
    "id": "1.0.109.0/24",
    "country": "JP",
    "state": "Hiroshima",
    "city": "Hiroshima",
    "zip": "732-0057",
    "timezone": "+09:00"
  },
  {
    "$index": {
      "cidr": "1.0.110.0/24"
    },
    "id": "1.0.110.0/24",
    "country": "JP",
    "state": "Shimane",
    "city": "Izumo",
    "zip": "693-0044",
    "timezone": "+09:00"
  }
]`, buf.String())
}
