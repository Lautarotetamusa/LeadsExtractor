const HOST = "http://localhost:8081";

window.onload = function() {
  var propiedadesInput = document.getElementById("propiedadesInput");
  propiedadesInput.value = "";
};

async function execute_script() {
  var error = false;

  var portal = document.getElementById("portalSelector").value;
  var nombreAsesor = document.getElementById("nombreAsesorInput").value;
  var emailAsesor = document.getElementById("emailAsesorInput").value;
  var telefonoAsesor = document.getElementById("telefonoAsesorInput").value;

  var propiedadesInput = document.getElementById("propiedadesInput");
  var propiedades = propiedadesInput.value;

  var errorNombreAsesor = document.getElementById("errorNombreAsesor");
  var errorEmailAsesor = document.getElementById("errorEmailAsesor");
  var errorTelefonoAsesor = document.getElementById("errorTelefonoAsesor");
  var errorPropiedades = document.getElementById("errorPropiedades");

  errorNombreAsesor.innerHTML = "";
  errorEmailAsesor.innerHTML = "";
  errorTelefonoAsesor.innerHTML = "";
  errorPropiedades.innerHTML = "";

  if (nombreAsesor.trim() === "") {
    errorNombreAsesor.innerHTML = "Por favor, ingresa el nombre del asesor.";
    error = true;
  }

  if (emailAsesor.trim() === "") {
    errorEmailAsesor.innerHTML = "Por favor, ingresa el email del asesor.";
    error = true;
  }

  if (telefonoAsesor.trim() === "") {
    errorTelefonoAsesor.innerHTML = "Por favor, ingresa el teléfono del asesor.";
    error = true;
  }

  if (propiedades.trim() !== "") {
    try {
      JSON.parse(propiedadesInput.value);
    } catch (SyntaxError) {
      errorPropiedades.innerHTML = "Las propiedades ingresadas no son un JSON válido.";
      error = true;
    }
  }

  if (error) {
    return;
  }

  // Mostrar la imagen de carga
  document.getElementById("loadingSpinner").style.display = "block";

  try {
    // Enviar los datos al servidor mediante la API Fetch
    const response = await fetch(`${HOST}/cotizacion`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        portal: portal,
        asesor: {
          name: nombreAsesor,
          email: emailAsesor,
          phone: telefonoAsesor
        },
        posts: propiedades ? JSON.parse(propiedades) : null
      })
    });

    const data = await response.json();
    const taskId = data.taskId;
    checkTaskStatus(taskId);
  } catch (error) {
    console.error(error);
    alert("Ocurrió un error iniciando la cotización");
    document.getElementById("loadingSpinner").style.display = "none";
  }
}

function checkTaskStatus(taskId) {
  const interval = setInterval(async () => {
    try {
      const response = await fetch(`${HOST}/check-cotizacion-status/${taskId}`);
      const data = await response.json();

      if (data.status === 'completed') {
        clearInterval(interval);
        document.getElementById("loadingSpinner").style.display = "none";
        alert('La cotización ha sido generada.');
        if (data.pdf_url) {
          window.open(HOST + "/" + data.pdf_url, '_blank');
        }
      } else if (data.status === 'error') {
        clearInterval(interval);
        document.getElementById("loadingSpinner").style.display = "none";
        alert('Ocurrió un error generando la cotización.');
      }
    } catch (error) {
      clearInterval(interval);
      console.error(error);
      alert("Ocurrió un error verificando el estado de la cotización");
      document.getElementById("loadingSpinner").style.display = "none";
    }
  }, 2000); // Revisa cada 2 segundos
}
