window.addEventListener('load', function(e){

    const list_controls = document.querySelector("#list-controls");
    const list_controls_models = document.querySelector("#list-controls-models");
    const list_controls_providers = document.querySelector("#list-controls-providers");        

    const list_model_select = document.querySelector("#list-model-select");    
    const list_provider_select = document.querySelector("#list-provider-select");
    
    if (! list_controls){
	console.warn("Missing list-controls");
	return;
    }

    if (! list_controls_models){
	console.warn("Missing list-controls-models");
	return;
    }

    if (! list_controls_providers){
	console.warn("Missing list-controls-providers");
	return;
    }

    if (! list_model_select){
	console.warn("Missing list-model-select");
	return;
    }

    if (! list_provider_select){
	console.warn("Missing list-provider-select");
	return;
    }

    const main = document.querySelector("#main");
    const list_uri = main.getAttribute("data-list-uri");
    
    list_controls_models.onchange = function(e){

	const el = e.target;
	const v = el.value;

	const u = new URL(list_uri, location);
	const s = new URLSearchParams();

	if (v != ""){
	    s.set("model", v);
	}

	if (list_provider_select.value != ""){
	    s.set("provider", list_provider_select.value);
	}

	u.search = s;
	location.href = u.toString();
	return false;
    };

    list_controls_providers.onchange = function(e){

	const el = e.target;
	const v = el.value;

	const u = new URL(list_uri, location);
	const s = new URLSearchParams();

	if (v != ""){
	    s.set("provider", v);
	}

	if (list_model_select.value != ""){
	    s.set("model", list_model_select.value);
	}

	u.search = s;
	
	location.href = u.toString();
	return false;	
    };

    list_controls.style.display = "block";
});
