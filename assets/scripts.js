$(document).ready(function(){
        // Add minus icon for collapse element which is open by default
        $(".collapse.in").each(function(){
        	$(this).siblings(".panel-heading").find(".glyphicon").addClass("glyphicon-minus").removeClass("glyphicon-plus");
        });
        
        // Toggle plus minus icon on show hide of collapse element
        $(".collapse").on('show.bs.collapse', function(){
        	$(this).parent().find(".glyphicon").removeClass("glyphicon-plus").addClass("glyphicon-minus");
        }).on('hide.bs.collapse', function(){
        	$(this).parent().find(".glyphicon").removeClass("glyphicon-minus").addClass("glyphicon-plus");
        });

        
        $('.resp-prettyprint').each(function() {
        	var ctx = $(this);
        	var html =  ctx.html();
		    ctx.html("")
		    
		    html = html.replaceAll('\\n', '');
		    html = html.replaceAll('\\t', '');
		    html = html.replaceAll('\\', '');
		    html = html.trim();

		    if (html.charAt(0) == '"'){
		    	html = html.substr(1);
		    	html = html.slice(0, -1);
		    }
		    var obj = JSON.parse(html);
			var formattedJson = JSON.stringify(obj, null, 4);
		    ctx.html("<pre>"+ syntaxHighlight(formattedJson) + "</pre>");
        });
        
        $.ajaxSetup({
            headers: {
                'X-Requested-With': 'xmlhttprequest',
            }
        });
        
    $("#btnSubmit").click(function (e) {
            // Get form
            var form = $('#frm_add_doc')[0];

            // Create an FormData object 
            var data = new FormData(form);

            e.preventDefault(); // avoid to execute the actual submit of the form.
            var form = $(this);
            $.ajax({
                type: "POST",
                url: "/upload-doc",
                enctype: 'multipart/form-data',
                data: data,
                processData: false,
                contentType: false,
                cache: false,
                timeout: 600000,
                // data: form.serialize(), // serializes the form's elements.
                success: function (resp) {
                    if (resp == 401) {
                        alert("Unauthorized! Please login");
                        return;
                    }
                    if (resp==200){
                        alert("API docs created successfully!"); // show response from the php script.
                        location.reload();
                        return;
                    }
                }
            });
        });

        $(".btn-delete-doc").click(function () {
            if (confirm("Are you sure!") == false){
                return;
            }
            var name = $(this).data('doc-name');
            $.post("/delete-doc", { name: name }, function (resp) {
                if (resp == 401) {
                    alert("Unauthorized! Please login");
                    return;
                }
                if (resp==200){
                   location.reload()
                    alert("Documentaion deleted!");
                }else{
                    alert("Failed to delete documentaion!");
                    return;
                }
            });
        });

 });


String.prototype.replaceAll = function (replaceThis, withThis) {
   var re = new RegExp(RegExp.quote(replaceThis),"g"); 
   return this.replace(re, withThis);
};


RegExp.quote = function(str) {
     return str.replace(/([.?*+^$[\]\\(){}-])/g, "\\$1");
};

function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}


