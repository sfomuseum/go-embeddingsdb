window.addEventListener("load", function load(event){

    const feedback = document.querySelector("#feedback");
    const input = document.querySelector("#input");
    const button = document.querySelector("#button");    
    
    sfomuseum.golang.wasm.fetch("wasm/oembeddings_validate.wasm").then((rsp) => {

	button.onclick = function(){

	    const oe = input.value;

	    oembeddings_validate(oe).then((rsp) => {
		feedback.innerText = "Document validates as OEmbeddings";
	    }).catch((err) => {
		console.error("Validation failed");
		feedback.innerText = "Validation failed: " + err;
	    });
	    
	    return false;
	};
	
	button.removeAttribute("disabled");
	
	
    }).catch((err) => {
	console.error("Failed to load WASM binary", err);
	feedback.innerText = "Failed to load WASM binary";
        return;
    });
    
});
