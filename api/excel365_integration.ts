async function loadDatabase(workbook: ExcelScript.Workbook) {
  const apiUrl = "https://55f2-2800-40-32-559-93af-6757-76a2-275b.ngrok-free.app/communication";
    try{
        const options = {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'ngrok-skip-browser-warning': '8080'
            }
        };
        const res = await fetch(apiUrl, options);

        console.log(res.status);

        if (!(res.ok)) {
            console.log("ERROR: no se pudo hacer la peticion a la api", res.status)
            return
        }

        const data: object = await res.json();
        const rows:    (string | boolean | number)[][] = data["rows"];
        const headers: (string | boolean | number)[][] = [data["headers"]];
        const rowCount: number  = data["row_count"];
        const colCount: number  = data["col_count"];

        let charCode: number = 'A'.charCodeAt(0) + colCount - 1;
        let endRange = String.fromCharCode(charCode); 
        const range: string = `A2:${endRange}${rowCount+1}`;

        console.log("range:", range)
        console.log("headers:", headers)

        // Obtener la hoja de cálculo activa
        const sheet = workbook.getActiveWorksheet();
        sheet.getRange(`A1:${endRange}1`).setValues(headers);
        sheet.getRange(range).setValues(rows);
    }catch (err) {
        console.log("ERROR: no se pudo hacer la peticion a la api", err)
        return
    }
}

async function updateAsesores(workbook: ExcelScript.Workbook) {
    const sheet = workbook.getActiveWorksheet();

    const values = sheet.getRange("A2:C99").getValues();

    let asesores: object[] = [];
    for (const row of values) {
        if (row[0] == "") {
            break;
        }

        asesores.push({
            "name": row[0],
            "phone": row[1].toString(),
            "active": row[2] == "Activo"
        });
    }
    console.log(asesores);

  const apiUrl = "https://reboraautomatizaciones.com/api/asesores/";
    try {
        const options = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'ngrok-skip-browser-warning': '8080'
            },
            body: JSON.stringify(asesores)
        };
        const res = await fetch(apiUrl, options);
        console.log(res.status);

        if (!(res.ok)) {
            const text = await res.text();
            console.log("ERROR: no se pudo hacer la peticion a la api", res.status, text)
            return
        }

        console.log("SUCCESS: peticion realizada con exito")
    } catch (err) {
        console.log("ERROR: no se pudo hacer la peticion a la api", err)
        return
    }
}
