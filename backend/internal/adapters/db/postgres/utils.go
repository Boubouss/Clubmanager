package postgres

import "strconv"


func generateUpdateQuery(table string, data map[string]string) (string, []any) {
  query := "UPDATE " + table + " SET "
  args := make([]any, 0)
  i := 1
  id := data["id"]
  delete(data, "id")

  for k, v := range data {
    args = append(args, v)
    query += k 
    query += " = $"
    query += strconv.Itoa(i)
    i++
    if i < len(data) {
      query += ", "
    } 
  }

  query += " WHERE id = $"
  query += strconv.Itoa(i)
  args = append(args, id)

  return query, args
}
