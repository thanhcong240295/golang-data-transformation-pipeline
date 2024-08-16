package transformation_data_core

import (
	"database/sql"
	"encoding/csv"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type CsvRecord struct {
	CatId         string
	Cat           string
	SubCatId      string
	SubCat        string
	ProductCode   string
	ProductName   string
	Msrp          int
	StandardPrice int
	Description   string
	ImageUrl      string
	OnStock       int
}

func openFile(filePath string) *os.File {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Opening csv file has an error")
	}

	return f
}

func readFile(file *os.File) [][]string {
	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	data, err := csvReader.ReadAll()

	if err != nil {
		log.Fatal().Err(err).Msg("Reading csv file has an error")
	}

	return data
}

func getProductList(data [][]string) []CsvRecord {
	var productList []CsvRecord
	for i, line := range data {
		if i > 0 {
			var rec CsvRecord
			for j, field := range line {
				if j == 0 {
					rec.CatId = field
				}

				if j == 1 {
					rec.Cat = field
				}

				if j == 2 {
					rec.SubCatId = field
				}

				if j == 3 {
					rec.SubCat = field
				}

				if j == 4 {
					rec.ProductCode = field
				}

				if j == 5 {
					rec.ProductName = field
				}

				if j == 6 {
					i, err := strconv.Atoi(field)
					if err != nil {
						rec.Msrp = 0
					} else {
						rec.Msrp = i
					}
				}

				if j == 7 {
					i, err := strconv.Atoi(field)
					if err != nil {
						rec.StandardPrice = 0
					} else {
						rec.StandardPrice = i
					}
				}

				if j == 8 {
					rec.Description = field
				}

				if j == 9 {
					rec.ImageUrl = field
				}

				if j == 10 {
					rec.OnStock = getOnStockValue(field)
				}
			}
			productList = append(productList, rec)
		}
	}

	return productList
}

func getOnStockValue(field string) int {
	if field == "True" {
		return 1
	}

	return 0
}

func generateCodeByName(name string) string {
	return strings.Replace(strings.ToLower(name), " ", "-", -1)
}

func getLastedOrderCat(db *sql.DB) string {
	var lastedDisplayOrder sql.NullString
	err := db.QueryRow("select max(c.display_order) from category c").Scan(&lastedDisplayOrder)

	if err != nil {
		log.Fatal().Err(err).Msg("Query lasted category order has an error")
	}

	order := "0"
	if lastedDisplayOrder.Valid {
		order = lastedDisplayOrder.String
	}

	i, _ := strconv.Atoi(order)
	i = i + 1

	return strconv.Itoa(i)
}

func upsertCategory(db *sql.DB, name string, parentName string) string {
	var parentCatId sql.NullString
	var catId sql.NullString

	parentCode := generateCodeByName(parentName)
	parentQuery, _ := db.Query("INSERT IGNORE INTO category(name, code, display_order, slug) VALUES (?, ?, ?, ?)", parentName, parentCode, getLastedOrderCat(db), parentCode)

	parentQuery.Close()

	parentErr := db.QueryRow("select id from category where name = ?", parentName).Scan(&parentCatId)

	if parentErr != nil {
		log.Fatal().Err(parentErr).Msg("Query parent category has an error")
	}

	err := db.QueryRow("select id from category where name = ?", parentName).Scan(&catId)

	if err != nil {
		log.Fatal().Err(err).Msg("Query category has an error")
	}

	code := strings.Replace(strings.ToLower(name), " ", "-", -1)
	query, _ := db.Query("INSERT IGNORE INTO category(name, code, display_order, slug, parent_id) VALUES (?, ?, ?, ?, ?)", name, code, getLastedOrderCat(db), code, parentCatId)

	query.Close()
	return catId.String
}

func generateZeroString(length int) string {
	const charset = "0"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}

func generateSku(db *sql.DB) string {
	var prefix string
	var currentNumber string
	var lengthLimit int
	var nextNumber int

	err := db.QueryRow("select sn.prefix, sn.next_number, sn.length_limit from sequence_number sn where sn.function_name = 'PRODUCT'").Scan(&prefix, &currentNumber, &lengthLimit)

	if err != nil {
		log.Fatal().Err(err).Msg("Query sequence number has an error")
	}

	i, _ := strconv.Atoi(currentNumber)
	nextNumber = i + 1
	updateQuery, _ := db.Query("update sequence_number sn set sn.next_number = ? where sn.function_name = 'PRODUCT'", nextNumber)

	updateQuery.Close()

	usedLength := len(prefix + currentNumber)

	remainingLength := lengthLimit - usedLength

	return prefix + generateZeroString(remainingLength) + currentNumber
}

func upsertProducts(db *sql.DB, productList []CsvRecord) {
	for _, p := range productList {
		catId := upsertCategory(db, p.SubCat, p.Cat)
		var productId sql.NullString

		db.QueryRow("SELECT id FROM product WHERE name = ?", p.ProductName).Scan(&productId)

		if productId.String == "" {
			productQuery, err := db.Query(`
				INSERT IGNORE INTO product(name, category_id, uom_id, description) 
				VALUES(?, ?, ?, ?)
				`,
				p.ProductName,
				catId,
				1,
				p.Description,
			)

			if err != nil {
				log.Fatal().Err(err).Msg("Insert product has an error")
			}

			productQuery.Close()

			db.QueryRow("SELECT id FROM product WHERE name = ?", p.ProductName).Scan(&productId)
		}

		var skuId sql.NullString
		db.QueryRow("SELECT id FROM sku WHERE name = ? AND product_id = ?", p.ProductName, productId).Scan(&skuId)

		if skuId.String == "" {
			skuQuery, skuErr := db.Query(`
					INSERT IGNORE INTO sku(sku, name, product_id, on_stock, quantity, standard_price, msrp)
					VALUES(?, ?, ?, ?, ?, ?, ?)
					`,
				generateSku(db),
				p.ProductName,
				productId,
				p.OnStock,
				9999,
				p.StandardPrice,
				p.Msrp,
			)

			if skuErr != nil {
				log.Fatal().Err(skuErr).Msg("Insert SKU has an error")
			}

			skuQuery.Close()
		}
	}
}

func Exec(file_path string, db *sql.DB) {
	file := openFile(file_path)
	data := readFile(file)

	productList := getProductList(data)

	upsertProducts(db, productList)

	defer file.Close()
	defer db.Close()
}
