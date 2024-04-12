const HOST = "http://localhost/app";
const PORT = 80;

window.onload = function(){
	var portalSelector = document.getElementById("portalSelector");
	var urlContainer = document.getElementById("urlContainer");
	var filtersContainer = document.getElementById("filtersContainer");

	var urlInput = document.getElementById("urlInput");
	var filtersInput = document.getElementById("filtersInput");
	var messageInput = document.getElementById("messageInput");
	
	toggleURLField();
};

function toggleURLField() {
	if (portalSelector.value === "inmuebles24" || portalSelector.value === "casasyterrenos") {
		urlContainer.style.display = "none";
		filtersContainer.style.display = "block";
	} else {
		urlContainer.style.display = "block";
		filtersContainer.style.display = "none";
	}

	urlInput.value = "";
	filtersInput.value = "";
	messageInput.value = "";
}

function execute_script() {
	var error = false;

	var portal = document.getElementById("portalSelector").value;
	var message = messageInput.value;

	var json_filters = portal === "inmuebles24" || portal === "casasyterrenos";
	var url = json_filters ? filtersInput.value : urlInput.value;

	var errorMessage = document.getElementById("errorMessage");
	var errorFilters = document.getElementById("errorFilters");
	var errorUrl = document.getElementById("errorUrl");
	errorMessage.innerHTML = ""; // Limpiar mensajes de error anteriores
	errorFilters.innerHTML = ""; // Limpiar mensajes de error anteriores
	errorUrl.innerHTML = ""; // Limpiar mensajes de error anteriores

	if (message.trim() === "") {
		errorMessage.innerHTML = "Por favor, ingresa un mensaje.";
		error = true;
	}

	if (url.trim() === "") {
		if (filtersContainer.style.display === "block"){
			console.log("no json filters");
			errorFilters.innerHTML = "Por favor, ingresa filtros.";
			error = true;
		}else{
			errorUrl.innerHTML = "Por favor, ingresa una url.";
			error = true;
		}
	}

	if (url.trim() !== "" && json_filters){
		try{
			JSON.parse(filtersInput.value);
		}catch(SyntaxError){
			errorFilters.innerHTML = "Los filtros ingresados no son un JSON valido";	
			error = true;
		}
	}

	console.log(error);
	if (error){
		return;
	}

	console.log("Portal:", portal);
	console.log("Mensaje:", message);
	console.log("URL o Filtros:", url);

	// Enviar los datos al servidor mediante la API Fetch
	fetch(`${HOST}:${PORT}/execute`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({
			portal: portal,
			message: message,
			url_or_filters: json_filters ? JSON.parse(url) : url
		})
	})
	.then(response => {
		response.text().then(text => {
			alert(text)
		});
	})
	.catch(error => {
		console.error(error);
		alert("Ocurrio un error ejecutando el script")
	});
}
