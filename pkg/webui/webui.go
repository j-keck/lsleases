package webui

import (
	"encoding/json"
	"github.com/j-keck/lsleases/pkg/cscom"
	"github.com/j-keck/plog"
	"net/http"
	"strings"
)

var log = plog.GlobalLogger()

type endpoint struct {
	path string
	hndl func(http.ResponseWriter, *http.Request)
}

type WebUI struct {
	endpoints []endpoint
}

func NewWebUI() WebUI {
	self := new(WebUI)
	self.registerEndpoints()
	return *self
}

func (self *WebUI) ListenAndServe(addr string) {
	browserAddr := addr
	if strings.HasPrefix(browserAddr, ":") {
		browserAddr = "http://localhost" + browserAddr
	} else {
		browserAddr = "http://" + browserAddr
	}

	log.Infof("startup webui on address: %s - you can open the webui at: %s",
		addr, browserAddr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Warnf("unable to start webui: %v", err)
	}
}

func (self *WebUI) registerEndpoints() {
	// webui
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(index))
	})

	// version
	http.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		if version, err := cscom.AskServer(cscom.GetVersion); err == nil {
			json, _ := json.Marshal(struct {
				V string `json:"version"`
			}{version.(cscom.Version).String()})
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		} else {
			log.Warnf("unable to lookup server version: %v", err)
		}
	})

	// leases listing
	http.HandleFunc("/api/leases", func(w http.ResponseWriter, r *http.Request) {
		var leases cscom.Leases
		var err error
		if since := r.URL.Query().Get("since"); len(since) != 0 {
			var resp cscom.ServerResponse
			resp, err = cscom.AskServerWithPayload(
				cscom.GetLeasesSince,
				since,
			)
			leases = resp.(cscom.Leases)
		} else {
			var resp cscom.ServerResponse
			resp, err = cscom.AskServer(cscom.GetLeases)
			leases = resp.(cscom.Leases)
		}

		if err == nil {
			json, _ := json.Marshal(leases)
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		} else {
			log.Warnf("unable to lookup leases: %v", err)
		}
	})

	// clear leases
	http.HandleFunc("/api/clear-leases", func(w http.ResponseWriter, r *http.Request) {
		cscom.TellServer(cscom.ClearLeases)
	})
}

var index = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <title>lsleases WebUI</title>
    <style>
      .container {
          display: flex;
          justify-content: center;
      }
      .content {
          flex-direction: column;
      }
      #notification {
          color: #b51515;
          margin-top: 10px;
          margin-left: 15px;
      }
      table {
          border-collapse: collapse;
      }
      th, td {
          padding: 1em;
          border-bottom: 1px solid #ddd;
      }
      tr:hover {
          background-color: #f5f5f5;
      }

      td#created {
          text-align: right;
      }
      td#ip {
          text-align: right;
      }
      #footer {
          margin-top: 5px;
          text-align: right;
          font-size: 11px;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div id="content">
        <div style="overflow-x:auto;">
          <div id="notification"></div>
          <table id="leases">
            <thead><tr><th>Captured</th><th>IP</th><th>MAC</th><th>Hostname</th></tr></thead>
            <tbody></tbody>
          </table>
        </div>
        <div id="footer">lsleases</div>
      </div>
    </div>

    <script language="javascript">
      let since = 0;

      get("/api/version", function(xhr) {
          let version = " v" + xhr.response.version;
          let node = document.createElement("span");
          node.appendChild(document.createTextNode(version));
          node.style = "font-size: 8px";
          document.getElementById("footer").appendChild(node);
      });

      window.setInterval(function() {
          get("/api/leases?since=" + since, function(xhr) {
              since = new Date().getTime() * 1000000;
              if(xhr.statusText == "OK") {
                  updateNotification("");
                  let leases = xhr.response;
                  leases.sort(function(a, b) { return a.Created > b.Created });
                  leases.forEach(function(item, index) {
                      updateLeasesTable(item);
                  });
              } else {
                  updateNotification("Unable to fetch leases");
              }
          })}, 1000);


        function get(path, cb) {
          let xhr;
          if (window.XMLHttpRequest) {
              xhr = new XMLHttpRequest();
          } else if (window.ActiveXObject) {
              xhr = new ActiveXObject("Microsoft.XMLHTTP");
          }
          if (!xhr) {
              updateNotification("cannot create an XMLHttp instance");
          }
          xhr.onreadystatechange = function() {
              if (xhr.readyState == XMLHttpRequest.DONE) {
                  cb(xhr);
              }
          };
          xhr.responseType = "json";
          xhr.open("GET", path);
          xhr.send();
      }

      function updateNotification(txt) {
          document.getElementById("notification").innerHTML = txt;
      }

      function updateLeasesTable(lease) {
          let tbody = document.getElementById("leases").getElementsByTagName("tbody")[0];

          // remove the old entry
          let old = document.getElementById(lease.Mac);
          if(old != null) {
              tbody.removeChild(old);
          }

          // create a new row
          let row = tbody.insertRow(0);
          row.id = lease.Mac;

          // convert timestamp
          let created = new Date(Date.parse(lease.Created));
          let createdCell = row.insertCell();
          createdCell.id = "created";
          createdCell.title = created.toLocaleString();
          createdCell.appendChild(document.createTextNode(created.toLocaleTimeString()));

          // text cells
          ["IP", "Mac", "Host"].forEach(function(n) {
              let cell = row.insertCell();
              cell.id = n.toLowerCase();
              cell.appendChild(document.createTextNode(lease[n]));
          });
      }
    </script>
  </body>
`
