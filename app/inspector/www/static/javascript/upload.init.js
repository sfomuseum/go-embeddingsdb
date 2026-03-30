window.addEventListener('load', function(e){

    const target = document.querySelector("#upload-similar");
    const upload_image = document.querySelector("#upload-image");
    const upload_input = document.querySelector("#upload");
    const model_input = document.querySelector("#upload-model-select");    
    
    const progress = document.querySelector("#progress");    
    const submit = document.querySelector("#submit");

    upload_input.addEventListener('change', function(){
	
	const file = this.files[0];

	if (!file) {
	    return;
	}

	const reader = new FileReader();

	reader.onload = function (e) {

	    const im = document.createElement("img");
	    im.src = e.target.result;

	    upload_image.innerHTML = "";
	    upload_image.appendChild(im);

	    // Result any existing similar results
	    target.innerHTML = "";	    
	};
	
	reader.readAsDataURL(file);
    });
    
    const display_similar = function(data){

	// TBD: Replace as much of this code as possible with HTML <template> elements?
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/template
	
	const count = data.length;
	
	const table_headers = [
	    "Distance",
	    "Provider",
	    "Depiction",
	    "Subject",
	    "Attributes",
	];
	
	const count_headers = table_headers.length;
	
	const grid_el = document.createElement("div");
	grid_el.setAttribute("class", "row similar-grid");
	
	for (var i=0; i < count; i++){
	    
	    const similar = data[i];
	    console.log("similar", similar);
	    
	    const image_el = document.createElement("div");
	    image_el.setAttribute("class", "similar-image");
	    
	    if (similar.attributes["uri"]){
		
		const img = document.createElement("img");
		img.setAttribute("src", similar.attributes["uri"]);
		image_el.appendChild(img);
	    }
	    
	    grid_el.appendChild(image_el);

	    // START OF properies (header)
	    
	    const table_el = document.createElement("table");
	    table_el.setAttribute("class", "table similar-table");
	    
	    const header_row = document.createElement("tr");
	    
	    for (var j=0; j < count_headers; j++){
		
		const header = document.createElement("th");
		header.appendChild(document.createTextNode(table_headers[j]));
		header_row.appendChild(header);
	    }
	    
	    table_el.appendChild(header_row);

	    // START OF properties (data)
	    
	    const similar_row = document.createElement("tr");
	    
	    const distance_cell = document.createElement("td");
	    distance_cell.appendChild(document.createTextNode(similar.similarity));
	    similar_row.appendChild(distance_cell);
	    
	    const provider_cell = document.createElement("td");
	    provider_cell.appendChild(document.createTextNode(similar.provider));
	    similar_row.appendChild(provider_cell);
	    
	    const depiction_cell = document.createElement("td");

	    const depiction_url = new URL("/", location);
	    const depiction_params = new URLSearchParams();

	    depiction_url.pathname = "/record/" + similar.provider + "/" + similar.depiction_id;
	    depiction_params.set("model", model_input.value);

	    depiction_url.search = depiction_params;
	    
	    const depiction_link = document.createElement("a");
	    depiction_link.setAttribute("href", depiction_url.toString());
	    
	    depiction_link.appendChild(document.createTextNode(similar.depiction_id));
	    depiction_cell.appendChild(depiction_link);
	    
	    similar_row.appendChild(depiction_cell);
	    
	    const subject_cell = document.createElement("td");
	    subject_cell.appendChild(document.createTextNode(similar.subject_id));
	    similar_row.appendChild(subject_cell);

	    // START OF attributes
	    
	    const attrs_cell = document.createElement("td");
	    
	    const attrs_table = document.createElement("table");
	    attrs_table.setAttribute("class", "table attributes");
	    
	    for (k in similar.attributes){
		
		const attr_row = document.createElement("tr");
		
		const attr_header = document.createElement("th");
		attr_header.appendChild(document.createTextNode(k));
		attr_row.appendChild(attr_header);
		
		const attr_cell = document.createElement("td");
		attr_cell.appendChild(document.createTextNode(similar.attributes[k]));
		attr_row.appendChild(attr_cell);
		
		attrs_table.appendChild(attr_row);
	    }
	    
	    attrs_cell.appendChild(attrs_table);
	    similar_row.appendChild(attrs_cell);
	    
	    // END OF attributes
	    
	    table_el.appendChild(similar_row);

	    // END OF properties (header and data)
	    
	    grid_el.appendChild(table_el);		
	}
	
	target.innerHTML = "";

	var summary;

	switch (count){
	    case 0:
		summary = "There are no matching records.";
		break;
	    case 1:
		summary = "There is one similar record.";
		break;
	    default:
		summary = "There are " + count + " similar records.";
		break;
	}

	const summary_el = document.createElement("div");
	summary_el.setAttribute("id", "summary");
	summary_el.appendChild(document.createTextNode(summary));

	target.appendChild(summary_el);
	target.appendChild(grid_el);
    };
    

    submit.onclick = function(e){

	const u = new URL("/api/upload/", location);
	const url = u.toString();
	
	const form = document.querySelector("#upload-form");
	const data = new FormData(form);

	const xhr = new XMLHttpRequest();
	xhr.open('POST', url);

	xhr.upload.addEventListener('progress', function (ev) {
	    
	    if (ev.lengthComputable) {
		const percent = Math.round((ev.loaded / ev.total) * 100);
		progress.value = percent;
		console.debug(`Upload progress: ${percent}%`);
	    }
	});

	xhr.addEventListener('load', function () {
	    
	    progress.style.display = 'none';
	    progress.value = 0;
	    
	    if (xhr.status != 200){
		console.error('Upload failed', xhr.status, xhr.statusText);
		return;
	    }

	    var data;
	    
	    try {
		data = JSON.parse(xhr.responseText);
	    } catch (err) {
		console.error("Failed to parse response", err);
		return;
	    }

	    display_similar(data);
	});

	target.innerHTML = "";
	progress.style.display = "inline-block";

	xhr.send(data);
	return false;
    };

    submit.removeAttribute("disabled");
});
