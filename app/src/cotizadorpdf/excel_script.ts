async function main(workbook: ExcelScript.Workbook) {
    try {
        // Obtener todas las hojas necesarias
        const datosVenta = workbook.getWorksheet("DatosVenta");
        const datosVendedor = workbook.getWorksheet("DatosVendedor");
        const datosFijos = workbook.getWorksheet("DatosFijos");

        datosVenta.getRange("C1").setValue("Generando...");
        datosVenta.getRange("C2").clear();
        
        // Verificar que todas las hojas existen
        if (!datosVenta || !datosVendedor || !datosFijos) {
            throw new Error("¡Faltan hojas en el libro! Verifica los nombres de las hojas.");
        }

        // Construir el payload desde las diferentes hojas
        const payload = {
            elaborado_por: {
                nombre: datosVendedor.getRange("B2").getValue() as string,
                telefono: datosVendedor.getRange("B3").getValue() as string,
                mail: datosVendedor.getRange("B4").getValue() as string,
                porcentaje_administracion: datosVendedor.getRange("B5").getValue() as number,
                is_valor_permisos: datosVendedor.getRange("B6").getValue() == "Si",
                valor_terreno: datosVendedor.getRange("B7").getValue() as number,
                area_terreno: datosVendedor.getRange("B8").getValue() as number
            },
            datos: {
                nombre: datosVenta.getRange("B2").getValue() as string
            },
            pagos: {
                inicial: datosVenta.getRange("B4").getValue() as number,
                porcentaje_inicio_obra: datosVenta.getRange("B5").getValue() as number,
                meses: datosVenta.getRange("B6").getValue() as number,
                tipo: datosVenta.getRange("B7").getValue() as string
            },
            areas_interiores: {
                sotano: datosVenta.getRange("B9").getValue() as number,
                planta_baja: datosVenta.getRange("B10").getValue() as number,
                planta_alta: datosVenta.getRange("B11").getValue() as number,
                roof: datosVenta.getRange("B12").getValue() as number,
                banos: datosVenta.getRange("B13").getValue() as number,
                cuartos: datosVenta.getRange("B14").getValue() as number,
            },
            areas_exteriores: {
                alberca: datosVenta.getRange("B16").getValue() as number,
                muro_perimetral: datosVenta.getRange("B17").getValue() as number,
                jardin: datosVenta.getRange("B18").getValue() as number,
                rampa: datosVenta.getRange("B19").getValue() as number,
            },
            valor_exteriores: {
                alberca: datosFijos.getRange("B2").getValue() as number,
                muro_perimetral: datosFijos.getRange("B3").getValue() as number,
                jardin: datosFijos.getRange("B4").getValue() as number,
                rampa: datosFijos.getRange("B5").getValue() as number,
            },
            valor_permisos: {
                licencia: datosFijos.getRange("B7").getValue() as number,
                gestorias: datosFijos.getRange("B8").getValue() as number,
                topografia: datosFijos.getRange("B9").getValue() as number,
                mecanica: datosFijos.getRange("B10").getValue() as number,
                calculo: datosFijos.getRange("B11").getValue() as number
            }
        };

        const host = "https://reboraautomatizaciones.com/app"
        const apiUrl = `${host}/generar_cotizacion`;

        // Enviar la solicitud POST
        const response = await fetch(apiUrl, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error(`Error en la API: ${response.status} ${response.statusText}`);
        }

        const resultado = await response.text();

        const link = `${host}/download/cotizacion${resultado}.html`;

         // Crear hipervínculo clickeable
        const linkCell = datosVenta.getRange("C2");
        datosVenta.getRange("C1").setValue("Cotizacion generada correctamente");
        linkCell.clear();
        linkCell.setValue("Abrir cotización");
        linkCell.getFormat().getFont().setBold(true);
        linkCell.setHyperlink({
            address: link,
            screenTip: "Haz clic para abrir la cotización PDF",
            textToDisplay: "Ver cotización PDF"
        });

        console.log("Cotización generada con éxito. ID: " + resultado);

    } catch (error) {
        // Manejo de errores mejorado
        const errorMessage = (error instanceof Error) ? error.message : "Error desconocido";

        // Mostrar error en Excel
        
        const datosVenta = workbook.getWorksheet("DatosVenta");
        datosVenta.getRange("C2").clear();
        datosVenta.getRange("C1").setValue(errorMessage);
        console.log("ERROR: " + errorMessage);
    }
}
