document.addEventListener('DOMContentLoaded', function() {
    const host = 'https://reboraautomatizaciones.com/api'; // Reemplaza con tu host real

    // Fetch para leads nuevos (is_new=true)
    fetchMetrics(`${host}/metrics?is_new=true`, 'newLeadsChart', 'Leads Nuevos', 'total-new-leads');

    // Fetch para leads duplicados (is_new=false)
    fetchMetrics(`${host}/metrics?is_new=false`, 'duplicateLeadsChart', 'Leads Duplicados', 'total-duplicate-leads');

    // Fetch para todos los leads (sin parámetros)
    fetchMetrics(`${host}/metrics`, 'allLeadsChart', 'Todos los Leads', 'total-all-leads');
});

function fetchMetrics(url, chartId, chartLabel, totalElementId) {
    fetch(url)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const totalLeads = data.data.total;
                document.getElementById(totalElementId).textContent = `Total: ${totalLeads}`;
                
                const dates = data.data.values.map(item => item.at);
                const counts = data.data.values.map(item => item.cant);
                
                createChart(chartId, chartLabel, dates, counts);
            } else {
                console.error('Error en la respuesta:', data);
            }
        })
        .catch(error => {
            console.error('Error al realizar la petición:', error);
            document.getElementById(totalElementId).textContent = 'Error al cargar el total';
        });
}

function createChart(chartId, chartLabel, labels, data) {
    const ctx = document.getElementById(chartId).getContext('2d');
    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: chartLabel,
                data: data,
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 2,
                fill: false,
                tension: 0.1
            }]
        },
        options: {
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Fecha'
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: 'Cantidad'
                    },
                    beginAtZero: true
                }
            }
        }
    });
}
