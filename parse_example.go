//go:build ignore

// run this using go run parse_example.go

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/iand/gtree"
	"github.com/kortschak/utter"
)

func main() {
	// 	input := `
	// 1.Henry Johnson  b: Abt. 1806 in Kilford, Ireland. d: 17 Sep 1861 in Swindon, Wiltshire, England; age: 55.
	//   + Alice O’Connor  b: Abt. 1800 in Limerick, Ireland. d: 12 Oct 1896 in Trowbridge, Wiltshire, England; age: 96.
	//   2.Elizabeth Johnson  b: 7 Dec 1838 in Chippenham, Wiltshire, England. d: Bef. 1928 in Swindon, Wiltshire, England; age: 89.
	//   + George Martin  b: abt 1835 in Ireland. m: 28 Jun 1857 in Swindon, Wiltshire, England. d: Mar 1883 in Swindon, Wiltshire, England; age: 48.
	//     3.Elizabeth Ann Martin  b: 24 Apr 1858. d: 1859; age: 0.
	//     3.Martha Martin  b: abt 1860 in Trowbridge, Wiltshire, England. d: Deceased.
	//   2.Susan Johnson  b: 25 Apr 1840 in Swindon, Wiltshire, England. d: Deceased.
	//   2.Anna Johnson  b: 22 May 1842 in Chippenham, Wiltshire, England. d: 1 Oct 1898 in Bath, Somerset, England; age: 56.
	//   + William Brown  b: Abt. 1839 in Limerick, Ireland. m: 13 Nov 1864 in Swindon, Wiltshire, England. d: 11 May 1867 in St. Luke’s Infirmary, Bath, Somerset, England; age: 28.
	//     3.Thomas Brown  b: 3 Nov 1865 in Trowbridge, Wiltshire, England. d: Deceased.
	//     + Charles Lewis  b: 1 Nov 1843 in Bristol, Gloucestershire, England. m: 7 Dec 1867 in Swindon, Wiltshire, England. d: Bef. 1871 in Trowbridge, Wiltshire, England; age: 27.
	//     3.Emily Lewis  b: 15 Oct 1868 in Swindon, Wiltshire, England. d: 8 Aug 1956 in Wiltshire, England; age: 87.
	//     + Alfred Green  b: 25 Feb 1864 in Norton, Somerset, England. m: 4 Sep 1888 in Swindon, Wiltshire, England. d: 28 Feb 1955 in Chippenham, Wiltshire, England; age: 91.
	//     + Joseph Navarro  b: 1840 in Bristol, Gloucestershire, England. m: 28 Oct 1872 in Swindon, Wiltshire, England. d: 15 July 1880 in Trowbridge, Wiltshire, England; age: 40.
	//   2.David Johnson  b: 15 Feb 1844 in Chippenham, Wiltshire, England. d: Oct 1916 in Swindon, Wiltshire, England; age: 72.
	//   + Martha Jane Harper  b: abt 1846 in Fleur-de-Lys, Monmouthshire, Wales. m: 17 Sep 1873 in St. Luke's Church, Swindon, Wiltshire, England. d: Jul 1923 in Swindon, Wiltshire, England; age: 77.
	//     3.John H Johnson  b: abt 1875 in Swindon, Wiltshire, England. d: Deceased.
	//     3.Jane Elizabeth Harper Johnson  b: 1880 in Swindon, Wiltshire, England. d: Deceased.
	//   2.James Johnson  b: 30 Mar 1849 in Chippenham, Wiltshire, England. d: 6 Apr 1849 in Chippenham, Wiltshire, England; age: 0.
	//   2.Peter Johnson  b: 2 Nov 1851 in Trowbridge, Wiltshire, England. d: Jun 1936 in Swindon, Wiltshire, England; age: 84.
	//   + Helen Clark  b: abt 1854 in Trowbridge, Wiltshire, England. m: 2 Dec 1872 in Christchurch, Wiltshire, England. d: 25 Jul 1935 in Swindon, Wiltshire, England; age: 81.
	//     3.Samuel Johnson  b: abt 1874 in Trowbridge, Wiltshire, England. d: Dec 1948 in Chippenham, Wiltshire, England; age: 74.
	//     + Mary Wells  b: abt 1875 in Nk, Wiltshire, England. m: Jul 1902 in Wiltshire, England. d: Deceased.
	//     3.Eliza Johnson  b: abt 1882 in Devizes, Wiltshire, England. d: Abt 1961 in Salisbury, Wiltshire, England; age: 79.
	// `

	input := `
1. O'Sullivan, Fiona (b. 1842-05-22 - Carmarthen, Carmarthenshire, Wales, d. 1898-10-01 - St. Luke's Infirmary, Swansea, Glamorgan, Wales)

  sp. Murphy, Sean (b. estimated 1839 - Limerick, Ireland, d. 1867-05-11 - St. Luke's Infirmary, Swansea, Glamorgan, Wales), m. 1864-11-11 - St. Andrew's Catholic Church, High Street, Swansea, Glamorgan, Wales

    2. Murphy, Michael (b. 1865-11-03 - Tredegar, Blaenau Gwent, Wales)

  sp. Bennett, Edward (b. 1843-11-01 - St. David's, Carmarthenshire, Wales, d. before 1871), m. 1867-12-07 - St. Andrew's Catholic Church, High Street, Swansea, Glamorgan, Wales

    2. Bennett, Mary (b. 1868-10-15 - Swansea, Glamorgan, Wales, d. 1956-08-08 - Glamorgan, Wales)

      sp. Powell, Arthur (b. 1864-02-25 - Bruton, Somerset, England, d. 1955-02-28 - Llanelli, Carmarthenshire, Wales), m. 1888-09-04 - Register Office, Swansea, Glamorgan, Wales

  sp. Cordova, Antonio (b. estimated 1840 - Spain, d. 1880-07-15 - North Colliery, Cwmbran, Torfaen, Wales), m. 1872-10-28 - St. Andrew's Catholic Church, High Street, Swansea, Glamorgan, Wales

    2. Brooks, Thomas (b. 1873-05-11 - Tredegar, Blaenau Gwent, Wales, d. 1916-12-02 - Maidstone, Kent, England)

      sp. Griffiths, Sarah (b. 1882-10-01 - Neath, Glamorgan, Wales, d. 1957-07-02 - Cwmbran, Torfaen, Wales), m. 1908-07-08 - St David's Church, Tredegar, Blaenau Gwent, Wales

    2. Brooks, Alice (b. 1876-08-02 - Pontypridd, Rhondda Cynon Taff, Wales, d. 1913-11-03 - 10 Hamilton Street, Gateshead, Tyne and Wear, England)

      sp. Wright, Henry (b. 1867-02-02 - Framlingham, Suffolk, England, d. 1923-12-09 - 34 Ormsby Street, Sunderland, Tyne and Wear, England), m. 1895-12-23 - Register Office, Swansea, Glamorgan, Wales

        3. Wright, Rebecca (b. 1897-01-02 - 14 College Street, Swansea, Glamorgan, Wales, d. 1984-12-03 - Gateshead, Tyne and Wear, England)

          sp. Ward, William (b. 1888-09-29 - Gateshead, Tyne and Wear, England, d. between 1954-04 and 1954-06 - Gateshead, Tyne and Wear, England), m. between 1916-10-01 and 1916-12-31 - Gateshead Registration District, Tyne and Wear, England

        3. Wright, Eileen (b. 1898-11-20 - The Barracks, Swansea, Glamorgan, Wales, d. 1949-11-07 - Gateshead, Tyne and Wear, England)

          sp. Murray, James (b. 1896-03-12 - Gateshead, Tyne and Wear, England, d. 1950-02-03 - Gateshead, Tyne and Wear, England), m. 1923-11-16 - St. Joseph's Roman Catholic Church, Sunderland, Tyne and Wear, England

        3. Wright, Hilda (b. 1901-04-05 - The Female Hospital, Royal Artillery Barracks, Woolwich, Kent, England, d. 1942-10-22 - 5 Windsor Road, Leeds, Yorkshire, England)

          sp. Baxter, Alfred (b. 1892-02-17 - Cleethorpes, Lincolnshire, England, d. between 1951-04 and 1951-06 - Sheffield Registration District, Yorkshire, England), m. between 1936-01-01 and 1936-03-31 - Doncaster Registration District, Yorkshire, England

        3. Wright, Edward (b. 1903-03-29 - Barracks, Hastings, Sussex, England, d. 1985-07-11 - Lewisham Registration District, London, England)

          sp. Kumar, Priya (b. 1905 - Pune, Maharashtra, British India, d. before 1978 - Pune, Maharashtra, British India), m. 1934-11-24 - Meerut, Uttar Pradesh, British India

          sp. Evans, Gwendolyn (b. 1911-05-17 - Dulwich, Surrey, England, d. 2003-02-06 - Bromley Registration District, London, England), m. 1969-07-18 - Lewisham Registration District, London, England

        3. Wright, Robert (b. 1905-08-26 - Naas, Kildare, Ireland, d. 1988-09-09 - Basildon Hospital, Basildon, Essex, England)

          sp. Lee, Florence (b. 1912-05-29 - 50 Elm Street, Ashington, Northumberland, England, d. 1996-06-03 - Fair Haven, Southend on Sea, Essex, England), m. 1935-04-20 - St. Paul's Catholic Church, Hammersmith, London, England

        3. Wright, Laura (b. 1907-07-10 - Naas, Kildare, Ireland, d. between 1979-10 and 1979-12 - Watford Registration District, Hertfordshire, England)

          sp. Collins, Victor (b. 1903-12-06 - Barnet Registration District, Middlesex, England, d. between 1984-01 and 1984-03 - Uxbridge Registration District, Middlesex, England), m. between 1935-10-01 and 1935-12-31 - Watford Registration District, Hertfordshire, England

        3. Wright, Helen Ann (b. 1909-04-29 - 34 Rose Street, Gateshead, Tyne and Wear, England, d. 2001-11 - Barnet Registration District, Greater London, England)

          sp. Hudson, Charles Edward (b. 1906-12-14 - Gateshead Registration District, Tyne and Wear, England, d. 1975-02-23 - Camden Registration District, London, England), m. 1937-09-28 - Gateshead Registration District, Tyne and Wear, England

        3. Wright, Joseph Patrick (b. 1911-10-18 - 10 Hamilton Street, Gateshead, Tyne and Wear, England, d. 1997-09-17 - Windsor, Berkshire, England)

          sp. Ellis, Catherine Mary Trevithick (b. 1912-12-12 - Watford, Hertfordshire, England, d. 1996-10-27 - Maidenhead, Berkshire, England), m. 1937-01-16 - Bexley, Kent, England

    2. Brooks, Rose (b. 1878-04-28 - Pontypridd, Rhondda Cynon Taff, Wales, d. 1964-06-11 - Swansea, Glamorgan, Wales)

      sp. Harper, Charles (b. estimated 1880 - South Molton, Devon, England, d. 1948-06-29 - Swansea, Glamorgan, Wales)

    2. Brooks, Margaret (b. 1880-07-17 - Tredegar, Blaenau Gwent, Wales, d. between 1973-07 and 1973-09 - Newport Registration District, Monmouthshire, Wales)

      sp. Brown, James Edward (b. estimated 1870 - Tredegar, Blaenau Gwent, Wales), m. 1901-09-11 - Register Office, Tredegar Registration District, Monmouthshire, Wales

  sp. Williams, David (b. 1851 - Blaenavon, Monmouthshire, Wales, d. 1941), m. 1889-12-16 - Register Office, Swansea, Glamorgan, Wales
`
	r := strings.NewReader(input)
	parser := new(gtree.Parser)

	chart, err := parser.Parse(context.Background(), r)
	if err != nil {
		fmt.Println("Error parsing input:", err)
		return
	}

	svg, err := gtree.SVG(chart.Layout(gtree.DefaultLayoutOptions()))
	if err != nil {
		fmt.Println("Error generating SVG:", err)
		return
	}

	fmt.Println(svg)
}
