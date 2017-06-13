/*
 * misc interactive stuff
 * interactive.js
 * Copyright 2012-2015
 * Laurence - Level 10 Fireball
 */

var site = location.protocol + '//' + location.host,
	cid  = 0,
	prev_new_link = "",
	user_list,
	ul_char,
    sel_c = [];

function id(el) {
    return document.getElementById(el);
}

function autoGrow(oField) {
    if (oField.scrollHeight > oField.clientHeight)
        oField.style.height = oField.scrollHeight + "px";
}

function textCounter() {
    var post = id('form_reply_form').elements['post'].value;
	if (post.length > char_limit_max)
        post = post.substring(0, char_limit_max);
    else
        id('forum_char_limit').innerHTML = (char_limit_max - post.length) + char_limit_str;
}

function postToCache(thread_id) {
    localStorage.setItem("p_post_t"+thread_id, id('form_reply_form').elements['post'].value);
    localStorage.setItem("p_pid_t"+thread_id, id('form_reply_form').elements['parent_id'].value);
}

function postFromCache(thread_id) {
    var post = localStorage.getItem("p_post_t"+thread_id),
		txt = id('form_reply_form').elements['post'];
    if (txt.value != "" || post == undefined || post == "") return;
    
    txt.value = post;
    parent_id = localStorage.getItem("p_pid_t"+thread_id);
	if (parent_id > 0) {
    	id('form_reply_form').elements['parent_id'].value = parent_id;
    	id('postcomment').innerHTML = "Reply to " + id('cu_'+parent_id).innerHTML + ":";
	}
	id('postcomment').scrollIntoView();
    txt.focus();
}

function postClear(thread_id) {
	id('postcomment').innerHTML = "Leave a comment";
	var txt = id('form_reply_form').elements['post'];
	txt.value = "";
	txt.focus();
	postRmCache(thread_id);
}

function postRmCache(thread_id) {
    localStorage.setItem("p_post_t"+thread_id,"");
    localStorage.setItem("p_pid_t"+thread_id,"");
}

function fieldHelper(oField, span, msg) {
    if (oField.value.length > 0)
		id(span).innerHTML = "";
    else
		id(span).innerHTML = msg;    
}

function addClass(c, className) {
	id(c).className += " "+className;
}

function delClass(c, className) {
	//id(c).className = id(c).className.replace(new RegExp("(^| )"+className+"( |$)", "g"), '');
	id(c).className = id(c).className.replace(new RegExp(className, "g"), '');
}

function clickComment(c) {
    if (sel_c[c] == true) {
        delClass('comment'+c,'cfocus');
        addClass('comment'+c,'cblur');
    } else {
        delClass('comment'+c,'cblur');
        addClass('comment'+c,'cfocus');
    }
    sel_c[c] = !sel_c[c];
}

function hideBranch(c) {
    addClass('nest'+c,'hide');
    delClass('nesth'+c,'hidden0');
}

function showBranch(c) {
    addClass('nesth'+c,'hidden0');
    delClass('nest'+c,'hide');
}

function replyTo(parent_id, q) {
    window.onscroll = undefined;
    if (id("cir_"+parent_id).innerHTML == 'Quote') {
        fetch("/ajax/quoteComment/"+parent_id, quoteComment);
    } else {
		id('reply_to_content').innerHTML = id('cc_'+parent_id).innerHTML;
		id('reply_quote').setAttribute('style', 'display: block;');
    }
    addClass('reply','c_reply_show');
    id('postcomment').innerHTML = "Reply to " + id('cu_'+parent_id).innerHTML + ":";
    id('form_reply_form').elements['parent_id'].value = parent_id;
    id('postcomment').scrollIntoView();
    id('form_reply_form').elements['post'].focus();
}

function quoteComment(s) {
    window.onscroll = undefined;
    var txt = id('form_reply_form').elements['post'];
    txt.value = s+txt.value;
    autoGrow(txt);
}

function showHiddenContainer(c) {
    id(c).setAttribute('style', 'display: block;');
	//delClass(c, 'hide');
    id(c).scrollIntoView();
}

function commentHideReply() {
    id('reply_header').setAttribute('style','display:none;');
    delClass('reply', 'c_reply_show');
    addClass('reply', 'c_reply_hide');
}

function gDisplayImg(gid, i) {
    addClass(gid+"_t"+i, "g_basic_t_selected");
    //id(gid+"_img").scrollIntoView();

    if ((window["i_"+gid]) != i) {
        delClass(gid+"_t"+(window["i_"+gid]), "g_basic_t_selected");
        (window["i_"+gid])=i;

        id(gid+"_img").src = (window[gid+"_path"])+(window[gid])[i];
        id(gid+"_a").href= (window[gid+"_path"])+"../"+(window[gid])[i];
    }
}

function gNav(gid, i) {
    var n = (window["i_"+gid]) + i,
    	l = (window[gid]).length-1;

    if (n > l) n = 0;
    if (n < 0) n = l;
    
    gDisplayImg(gid, n);
}

function giveKarma(c) {
	fetch("/ajax/giveKarma?c="+c+"&t="+token, commentKarma);
	cid=c;
}

function commentKarma(s) {
	id('ci_k'+cid).innerHTML = s;
}

function returnUserList(s) {
	if (s == undefined || s == "" ) return;
	user_list = JSON.parse(s);
	updateUserList();
}

function updateUserList() {
	function ul_used(ul) {
		for (var j=0; j<uls.length; j++) {
			if (uls[j] == ul.ID) return true;
		}
		return false;
	}
	
    var s = id("ul_input").value.toLowerCase(),
		l = s.length,
		c = s.substr(0, 1),
		list = "";

	if (c != ul_char) {
		ul_char = c;
		fetch("/ajax/userList/"+c+"?t="+token, returnUserList);
	}
	
	if (user_list == undefined) return;
	
	if (l == 0) {
		id('ul_users').innerHTML = "";
		return;
	}
	
	for (var i = 0; i < user_list.length; i++) {
		if (user_list[i].Alias.substr(0, l).toLowerCase() == s || user_list[i].Name.substr(0, l).toLowerCase() == s) {
			var ul = user_list[i];
			if (!ul_used(ul)) {
				list += '<a href="javascript:selectUserList('+ul.ID+');">'+ul.Alias+" ("+ul.Name+")</a><br/>";
			}
		}
	}
	id('ul_users').innerHTML = list;
}

function selectUserList(uid) {
	for (var i=0; i<user_list.length; i++) {
		if (user_list[i].ID == uid) {
			var ul = user_list[i];
			for (var j=0; j<uls.length; j++) {
				if (uls[j] == ul.ID) return;
			}
			
			uls.push(ul.ID);
			formUserList();
			id("ul_to").innerHTML += '<div id="ul_id'+ul.ID+'" class="ul_item nowrap hide" onclick="removeUserList('+ul.ID+');">'+ul.Alias+'</div>';
			id("ul_input").value = "";
			id("ul_input").focus();
			setTimeout(function() { delClass("ul_id"+ul.ID, "hide"); addClass("ul_id"+ul.ID, "show"); }, 250);
			updateUserList();
			return;
        }
	}
}

function removeUserList(uid) {	
	id("ul_input").focus();
	for (var i=0; i<uls.length; i++) {
		if (uls[i] == uid) {
			delClass("ul_id"+uid, "show");
			addClass("ul_id"+uid, "hide");
			setTimeout(function() { var e=id('ul_id'+uid); e.parentNode.removeChild(e); uls.splice(i, 1); formUserList(); }, 1000);
			return;
		}
	}
}

function formUserList() {
	id("ul_form_to").value = "";
	for (var i=0; i<uls.length; i++) {
		id("ul_form_to").value += uls[i]+",";
	}
}

function urlPreview() {
	var url = id('form_reply_form').elements['linkurl'];
	if (url.value == "") return;
	if (!url.value.match(/^[a-zA-Z0-9]+:\/\//)) url.value = "http://"+url.value;
		
	var s = url.value;
	if (prev_new_link != s) {
    	prev_new_link = s;
		fetch("/ajax/urlPreview?u="+encodeURIComponent(s)+"&t="+token, returnPreview);
	}
}

function returnPreview(s) {
    var json = JSON.parse(s);
    frm = id('form_reply_form')
    frm.elements['title'].value = json.Title;
    frm.elements['post'].value = json.Desc;
	frm.elements['mime'].value = json.Mime;
    if (json.Err != undefined) {
		delClass('parse_err_container', 'hide');
		id('parse_err').innerHTML = json.Err;
	} else {
		addClass('parse_err_container', 'hide');
	}
    
    autoGrow(frm.elements['post']);
    textCounter();
}

// stops events from cascading
function cancelBubble(e) {
    var event = e || window.event;
    if (typeof event.stopPropagation != "undefined")
        event.stopPropagation();
    else if (typeof event.cancelBubble != "undefined")
        event.cancelBubble = true;
}

function isElInView(el) {
	if (el == null) return;
	var rect = el.getBoundingClientRect();
	
	return (
		rect.top >= 0 &&
		rect.left >= 0 &&
		rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
		rect.right <= (window.innerWidth || document.documentElement.clientWidth)
	);
}

function detectBottom() {
    if (page >= max_pages) return;

	if (isElInView(id('t_page'+page))) {
        window.onscroll = undefined;
        fetch("/ajax/threadNextPage/"+tid+"/"+(page+1)+"/"+t+"/"+tname, threadNextPage);
    }
}

function threadNextPage(s) {
    if (s == "") return;
    
	delClass('t_page'+page,'t_page');
	id('t_page'+page).innerHTML += '<p>&nbsp;</p><div class="clear"></div><hr/>' + s;
    page++;
	
	
	var doc_loc = document.location.href.replace(/[\?\#].*$/g, "");
	doc_loc = doc_loc.replace(/\/+$/, "");
	doc_loc = doc_loc.replace(/\/page[0-9]+$/, "");
	history.pushState(null, null, doc_loc+"/page"+page);
	
    if (page < max_pages) window.onscroll = function() { detectBottom(); };
}

///////////////////////////////////////////////////////////////////////////////

function fetch(url, f) {   
    if (window.XMLHttpRequest) {
        // native XMLHttpRequest object
        req = new XMLHttpRequest();
    } else if (window.ActiveXObject) {
        // IE/Windows ActiveX version
        req = new ActiveXObject("Microsoft.XMLHTTP");
    } else {
        return 0;
    }

    req.onreadystatechange = function() { fetchHandler(f); };
    req.open("GET", site + url, true);
    req.send();
}
 
function fetchHandler(f) {
    // only if req is "loaded"
    if (req.readyState == 4) {
        // only if "OK"
        if (req.status == 200 || req.status == 304) {
            f(req.responseText);
        } else {
            //"ajax error (" + req.status + "):" + req.statusText;
            f("");
        }
    }
}