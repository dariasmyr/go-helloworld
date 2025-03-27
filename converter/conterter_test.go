package converter

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

// Define the structure of the XML data
type Transactions struct {
	XMLName      xml.Name      `xml:"TRANSACTIONS"`
	Transactions []Transaction `xml:"TRANSACTION"`
}

type Transaction struct {
	DELETE struct {
		SOCCDE     string `xml:"SOCCDE"`
		SOCWORKCDE string `xml:"SOCWORKCDE"`
	} `xml:"DELETE"`
	NEW struct {
		WORK struct {
			CPN struct {
				CPNIP struct {
					FSTNAME  string `xml:"FSTNAME"`
					IPNAMENR string `xml:"IPNAMENR"`
					LSTNAME  string `xml:"LSTNAME"`
				} `xml:"CPNIP"`
				CPNIPS   int    `xml:"CPNIPS"`
				CPNISWC  string `xml:"CPNISWC"`
				CPNTTL   string `xml:"CPNTTL"`
				DUR      string `xml:"DUR"`
				NATDBKEY struct {
					SOCCDE     string `xml:"SOCCDE"`
					SOCWORKCDE string `xml:"SOCWORKCDE"`
					SRCDB      string `xml:"SRCDB"`
				} `xml:"NATDBKEY"`
			} `xml:"CPN"`
			CPNS       int    `xml:"CPNS"`
			CPRDT      string `xml:"CPRDT"`
			CPRNR      int    `xml:"CPRNR"`
			CPSTTYP    int    `xml:"CPSTTYP"`
			CRETS      string `xml:"CRETS"`
			CREUID     int    `xml:"CREUID"`
			CUTNR      int    `xml:"CUTNR"`
			POSTDT     string `xml:"POSTDT"`
			PRDTITLE   int    `xml:"PRDTITLE"`
			SHARETISS  int    `xml:"SHARETISS"`
			SOCCDE     string `xml:"SOCCDE"`
			SOCWORKCDE string `xml:"SOCWORKCDE"`
			TXTMUSREL  int    `xml:"TXTMUSREL"`
			UPDTS      string `xml:"UPDTS"`
			UPUID      int    `xml:"UPUID"`
		} `xml:"WORK"`
	} `xml:"NEW"`
	UPDATE struct {
		WORK struct {
			CPN struct {
				CPNIP struct {
					FSTNAME  string `xml:"FSTNAME"`
					IPNAMENR string `xml:"IPNAMENR"`
					LSTNAME  string `xml:"LSTNAME"`
				} `xml:"CPNIP"`
				CPNIPS   int    `xml:"CPNIPS"`
				CPNISWC  string `xml:"CPNISWC"`
				CPNTTL   string `xml:"CPNTTL"`
				DUR      string `xml:"DUR"`
				NATDBKEY struct {
					SOCCDE     string `xml:"SOCCDE"`
					SOCWORKCDE string `xml:"SOCWORKCDE"`
					SRCDB      string `xml:"SRCDB"`
				} `xml:"NATDBKEY"`
			} `xml:"CPN"`
			CPNS       int    `xml:"CPNS"`
			CPRDT      string `xml:"CPRDT"`
			CPRNR      int    `xml:"CPRNR"`
			CPSTTYP    int    `xml:"CPSTTYP"`
			CRETS      string `xml:"CRETS"`
			CREUID     int    `xml:"CREUID"`
			CUTNR      int    `xml:"CUTNR"`
			POSTDT     string `xml:"POSTDT"`
			PRDTITLE   int    `xml:"PRDTITLE"`
			SHARETISS  int    `xml:"SHARETISS"`
			SOCCDE     string `xml:"SOCCDE"`
			SOCWORKCDE string `xml:"SOCWORKCDE"`
			TXTMUSREL  int    `xml:"TXTMUSREL"`
			UPDTS      string `xml:"UPDTS"`
			UPUID      int    `xml:"UPUID"`
		} `xml:"WORK"`
	} `xml:"UPDATE"`
}

func TestConvertXMLToCSV(t *testing.T) {
	t.Run("Convert XMT to CSV", func(t *testing.T) {
		// Open the XML file
		file, err := os.Open("data/transactions.xml")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			fmt.Println("Error getting file info:", err)
			return
		}
		if fi.Size() == 0 {
			fmt.Println("Error: XML file is empty")
			return
		}

		// Parse the XML file
		var transactions Transactions
		decoder := xml.NewDecoder(file)
		err = decoder.Decode(&transactions)
		if err != nil {
			fmt.Println("Error decoding XML:", err)
			return
		}

		// Create a CSV file
		csvFile, err := os.Create("data/transactions.csv")
		if err != nil {
			fmt.Println("Error creating CSV file:", err)
			return
		} else {
			fmt.Println("Document successfully decoded")
		}
		defer csvFile.Close()

		// Initialize CSV writer
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()
		writer.Comma = ';'

		// Write the CSV headers
		headers := []string{"SOCCDE", "SOCWORKCDE", "FSTNAME", "IPNAMENR", "LSTNAME", "CPNIPS", "CPNISWC", "CPNTTL", "DUR", "CPNS", "CPRDT", "CPRNR", "CPSTTYP", "CRETS", "CREUID", "CUTNR", "POSTDT", "PRDTITLE", "SHARETISS", "TXTMUSREL", "UPDTS", "UPUID"}
		writer.Write(headers)

		// Write the data from the parsed XML
		for _, transaction := range transactions.Transactions {
			// Extract data for NEW
			newData := transaction.NEW.WORK
			newRow := []string{
				newData.SOCCDE, newData.SOCWORKCDE,
				newData.CPN.CPNIP.FSTNAME, newData.CPN.CPNIP.IPNAMENR, newData.CPN.CPNIP.LSTNAME,
				fmt.Sprintf("%d", newData.CPN.CPNIPS), newData.CPN.CPNISWC, newData.CPN.CPNTTL, newData.CPN.DUR,
				fmt.Sprintf("%d", newData.CPNS), newData.CPRDT, fmt.Sprintf("%d", newData.CPRNR),
				fmt.Sprintf("%d", newData.CPSTTYP), newData.CRETS, fmt.Sprintf("%d", newData.CREUID),
				fmt.Sprintf("%d", newData.CUTNR), newData.POSTDT, fmt.Sprintf("%d", newData.PRDTITLE),
				fmt.Sprintf("%d", newData.SHARETISS), fmt.Sprintf("%d", newData.TXTMUSREL), newData.UPDTS,
				fmt.Sprintf("%d", newData.UPUID),
			}
			writer.Write(newRow)

			// Extract data for UPDATE (if applicable)
			updateData := transaction.UPDATE.WORK
			updateRow := []string{
				updateData.SOCCDE, updateData.SOCWORKCDE,
				updateData.CPN.CPNIP.FSTNAME, updateData.CPN.CPNIP.IPNAMENR, updateData.CPN.CPNIP.LSTNAME,
				fmt.Sprintf("%d", updateData.CPN.CPNIPS), updateData.CPN.CPNISWC, updateData.CPN.CPNTTL, updateData.CPN.DUR,
				fmt.Sprintf("%d", updateData.CPNS), updateData.CPRDT, fmt.Sprintf("%d", updateData.CPRNR),
				fmt.Sprintf("%d", updateData.CPSTTYP), updateData.CRETS, fmt.Sprintf("%d", updateData.CREUID),
				fmt.Sprintf("%d", updateData.CUTNR), updateData.POSTDT, fmt.Sprintf("%d", updateData.PRDTITLE),
				fmt.Sprintf("%d", updateData.SHARETISS), fmt.Sprintf("%d", updateData.TXTMUSREL), updateData.UPDTS,
				fmt.Sprintf("%d", updateData.UPUID),
			}
			writer.Write(updateRow)
		}

		fmt.Println("CSV file created successfully.")
	})
}
