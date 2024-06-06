const HOST = "http://localhost:8080";

window.onload = function(){
    var actions = document.getElementsByClassName("action");

	var actionsContainer = document.getElementById("actions-container");
    var actionHtml = actions[0].outerHTML;

    document.getElementById("add-action-btn").onclick = function (){
        actionsContainer.innerHTML += actionHtml;
    };

    document.getElementById("save-action-btn").onclick = function (){
        let jsonActions = [];
        for (const a of actions){
            jsonActions.push({
                "action": a.getElementsByTagName("select")[0].value,
                "interval": a.getElementsByTagName("input")[0].value
            });
        }

        console.log(jsonActions);
        saveAction(jsonActions);
    };
};

function saveAction(actions) {
	// Enviar los datos al servidor mediante la API Fetch
	fetch(`${HOST}/actions`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify([{
            condition: {
                "is_new": true
            },
            actions: actions
		}])
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
