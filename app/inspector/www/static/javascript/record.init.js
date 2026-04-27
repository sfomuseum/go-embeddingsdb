window.addEventListener('load', function(e){

    const model_current = document.querySelector("#model-current");
    const model_select = document.querySelector("#model-select");
    const record_table = document.querySelector("#record-table");
    
    const current_provider = record_table.getAttribute("data-provider");
    const current_depiction_id = record_table.getAttribute("data-depiction-id");
    const current_model = record_table.getAttribute("data-model");        

    const similar_controls = document.querySelector("#similar-controls");
    const model_provider = document.querySelector("#model-provider");

    const max_distance = document.querySelector("#max-distance");
    const max_distance_wrapper = document.querySelector("#max-distance-wrapper");
    const custom_max_distance = document.querySelector("#custom-max-distance");    

    const record_summary = document.querySelector("#record-summary");    
    const record_details = document.querySelector("#record-details");

    record_details.addEventListener("toggle", function(){

	if (record_details.open){
	    record_summary.style.display = "none";
	} else {
	    record_summary.style.display = "block";
	}
	
	return false;
    });
    
    custom_max_distance.addEventListener("change", function(){

	if (custom_max_distance.checked){
	    max_distance_wrapper.style.display = "block";
	} else {
	    max_distance_wrapper.style.display = "none";
	}
	
	return false;	
    });
    
    max_distance.addEventListener("input", function() {
	const el = document.querySelector("#max-distance-value");
	el.textContent = max_distance.value;	
    });
    
    
    const main = document.querySelector("#main");
    const record_uri = main.getAttribute("data-record-uri");
    
    const refine_btn = document.querySelector("#refine");

    refine_btn.onclick = function(){

	const u = new URL("/", location);
	const s = new URLSearchParams();
	
	u.pathname = record_uri + current_provider + "/" + current_depiction_id;

	if (model_select.value != ""){
	    s.set("model", model_select.value)
	}
	
	if (model_provider.value != ""){
	    s.set("similar-provider", model_provider.value);
	}

	if (custom_max_distance.checked){
	    s.set("custom-max-distance", "true");
	    s.set("max-distance", max_distance.value);
	}
	
	u.search = s;
	const href = u.toString();

	console.log("HREF", href);
	location.href= href;
	return false;
	
    }

    const u = new URL(location.href);

    if (u.searchParams.has("custom-max-distance")){
	custom_max_distance.checked = true;
	max_distance_wrapper.style.display = "block";
    }
    
    similar_controls.style.display = "block";
});
