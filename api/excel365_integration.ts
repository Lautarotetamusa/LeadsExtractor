async function main(workbook: ExcelScript.Workbook) {
    const apiUrl = "https://42b9-2800-40-32-559-93af-6757-76a2-275b.ngrok-free.app/communication";
    try{
        const options = {
            headers: {
                'Content-Type': 'application/json'
            }
        };
        const res = await fetch(apiUrl, options);

        console.log(res.status)
        console.log(res.text());

        if (!(res.ok)) {
            console.log("ERROR: no se pudo hacer la peticion a la api", res.json())
            return
        }

        const data: object = await res.json();
        console.log(data);
        const rows: (string | boolean | number)[][] = data["rows"];
        const rowCount: number = data["row_count"];
        const colCount: number = data["col_count"];
        const range: string = `A1:${'Z' + colCount}${rowCount}`;

        console.log("rows:", rows)

        const sheet = workbook.getActiveWorksheet();
        sheet.ge
    }catch (err) {
        console.log("ERROR: no se pudo hacer la peticion a la api", err)
        return
    }
}
