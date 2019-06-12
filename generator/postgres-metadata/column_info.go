package postgres_metadata

import (
	"database/sql"
	"fmt"
	"github.com/serenize/snaker"
	"strings"
)

type ColumnInfo struct {
	Name       string
	IsNullable bool
	DataType   string
	EnumName   string
}

func (c ColumnInfo) SqlBuilderColumnType() string {
	switch c.DataType {
	case "boolean":
		return "Bool"
	case "smallint", "integer", "bigint":
		return "Integer"
	case "date":
		return "Date"
	case "timestamp without time zone":
		return "Timestamp"
	case "timestamp with time zone":
		return "Timestampz"
	case "time without time zone":
		return "Time"
	case "time with time zone":
		return "Timez"
	case "USER-DEFINED", "text", "character", "character varying", "bytea", "uuid",
		"tsvector", "bit", "bit varying", "money", "json", "jsonb", "xml", "point", "interval", "line", "ARRAY":
		return "String"
	case "real", "numeric", "decimal", "double precision":
		return "Float"
	default:
		fmt.Println("Unknown sql type: " + c.DataType + ", using string column instead for sql builder.")
		return "String"
	}
}

func (c ColumnInfo) GoBaseType() string {
	switch c.DataType {
	case "USER-DEFINED":
		return snaker.SnakeToCamel(c.EnumName)
	case "boolean":
		return "bool"
	case "smallint":
		return "int16"
	case "integer":
		return "int32"
	case "bigint":
		return "int64"
	case "date", "timestamp without time zone", "timestamp with time zone", "time with time zone", "time without time zone":
		return "time.Time"
	case "bytea":
		return "[]byte"
	case "text", "character", "character varying", "tsvector", "bit", "bit varying", "money", "json", "jsonb",
		"xml", "point", "interval", "line", "ARRAY":
		return "string"
	case "real":
		return "float32"
	case "numeric", "decimal", "double precision":
		return "float64"
	case "uuid":
		return "uuid.UUID"
	default:
		fmt.Println("Unknown sql type: " + c.DataType + ", " + c.EnumName + ", using string instead for model type.")
		return "string"
	}
}

func (c ColumnInfo) GoModelType() string {
	typeStr := c.GoBaseType()
	if c.IsNullable && !strings.HasPrefix(typeStr, "[]") {
		return "*" + typeStr
	}

	return typeStr
}

func (c ColumnInfo) GoModelTag(isPrimaryKey bool) string {
	tags := []string{}

	if isPrimaryKey {
		tags = append(tags, "primary_key")
	}

	if len(tags) > 0 {
		return "`sql:\"" + strings.Join(tags, ",") + "\"`"
	}

	return ""
}

func getColumnInfos(db *sql.DB, dbName, schemaName, tableName string) ([]ColumnInfo, error) {

	query := `
SELECT column_name, is_nullable, data_type, udt_name
FROM information_schema.columns
where table_catalog = $1 and table_schema = $2 and table_name = $3
order by ordinal_position;`

	rows, err := db.Query(query, dbName, schemaName, tableName)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []ColumnInfo{}

	for rows.Next() {
		columnInfo := ColumnInfo{}
		var isNullable string
		err := rows.Scan(&columnInfo.Name, &isNullable, &columnInfo.DataType, &columnInfo.EnumName)

		columnInfo.IsNullable = isNullable == "YES"

		if err != nil {
			return nil, err
		}

		ret = append(ret, columnInfo)
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return ret, nil
}
