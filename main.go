package main

import (
   "net/http"
   "fmt"
   "time"
   "html/template"
   "io/ioutil"
   "strconv"
   "strings"
   "bytes"
   "log"
   "os/exec"
   "encoding/json"
)
//CPU'S
var cpus [5]string;
type CountG struct{
  counter string
}

type Ram struct {
  ramGraph string
}

//Create a struct that holds information to be displayed in our HTML file
type Welcome struct {
   Name string
   Time string
}

type User struct {
   Name string
   Password string
}

type Process struct {
    pid int
    cpu float64
}

var ram Ram

//Go application entrypoint   
func main() {
   //Instantiate a Welcome struct object and pass in some random information. 
   //We shall get the name of the user as a query parameter from the URL
   welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp)}
   profile :=   User{"Anonymous", "Anonymous"}
   mux := http.NewServeMux()
   mux.HandleFunc("/ram", ramPage)
   mux.HandleFunc("/cpu", cpuPage)
   mux.HandleFunc("/receive", receiveAjax)
   mux.HandleFunc("/receive2", receiveAjax2)
   mux.HandleFunc("/countProcess", countProcessURL)
   //We tell Go exactly where we can find our html file. We ask Go to parse the html file (Notice
   // the relative path). We wrap it in a call to template.Must() which handles any errors and halts if there are fatal errors
   
   templates := template.Must(template.ParseFiles("templates/index.html", "templates/process.html", "templates/cpu.html", "templates/ram.html"))

   //Our HTML comes with CSS that go needs to provide when we run the app. Here we tell go to create
   // a handle that looks in the static directory, go then uses the "/static/" as a url that our
   //html can refer to when looking for our css and other files. 
   
   mux.Handle("/static/", //final url can be anything
      http.StripPrefix("/static/",
         http.FileServer(http.Dir("static")))) //Go looks in the relative "static" directory first using http.FileServer(), then matches it to a
         //url of our choice as shown in http.Handle("/static/"). This url is what we need when referencing our css files
         //once the server begins. Our html code would therefore be <link rel="stylesheet"  href="/static/stylesheet/...">
         //It is important to note the url in http.Handle can be whatever we like, so long as we are consistent.

   //This method takes in the URL path "/" and a function that takes in a response writer, and a http request.
   mux.HandleFunc("/" , func(w http.ResponseWriter, r *http.Request) {

      //Takes the name from the URL query e.g ?name=Martin, will set welcome.Name = Martin.
      if name := r.FormValue("name"); name != "" {
         welcome.Name = name;
      }
      //If errors show an internal server error message
      //I also pass the welcome struct to the welcome-template.html file.      
      if err := templates.ExecuteTemplate(w, "index.html", welcome); err != nil {
         http.Error(w, err.Error(), http.StatusInternalServerError)
      }
   })

   //This method is to login
   // Login
   mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
      r.ParseForm()
 
      user := r.FormValue("user")  // Data from the form
      pwd := r.FormValue("password") // Data from the form
 
      dbPwd := "admin"
      dbUser := "admin"
 
      if user == dbUser && pwd == dbPwd {
         //fmt.Fprintln(w, "Login succesful!")
         profile.Name = dbUser
         profile.Password = dbPwd
         http.Redirect(w, r, "/process", http.StatusSeeOther)
         
      } else {
         //fmt.Fprintln(w, "Login failed!")
         welcome.Name = "Anonymous your user is Incorrect"
         http.Redirect(w, r, "/", http.StatusSeeOther) 
      }
    })

   mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
         if err := templates.ExecuteTemplate(w, "process.html", nil); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
         }
   })

   /*mux.HandleFunc("/cpu", func(w http.ResponseWriter, r *http.Request) {
         if err := templates.ExecuteTemplate(w, "cpu.html", nil); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
         }
      
   })*/

   /*mux.HandleFunc("/ram", func(w http.ResponseWriter, r *http.Request) {
         if err := templates.ExecuteTemplate(w, "ram.html", nil); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
         }
      
   })*/
   //Start the web server, set the port to listen to 8080. Without a path it assumes localhost
   //Print any errors from starting the webserver using fmt


   fmt.Println("Listening");
   fmt.Println(http.ListenAndServe(":8080", mux));
}

func getCPUSample() (idle, total uint64) {
    contents, err := ioutil.ReadFile("/proc/stat")
    if err != nil {
        return
    }
    lines := strings.Split(string(contents), "\n")
    for _, line := range(lines) {
        fields := strings.Fields(line)
        if fields[0] == "cpu" {
            numFields := len(fields)
            for i := 1; i < numFields; i++ {
                val, err := strconv.ParseUint(fields[i], 10, 64)
                if err != nil {
                    fmt.Println("Error: ", i, fields[i], err)
                }
                total += val // tally up all the numbers to get total ticks
                if i == 4 {  // idle is the 5th field in the cpu line
                    idle = val
                }
            }
            return
        }
    }
    return
}

func getRAMSample() string {
   contents, err := ioutil.ReadFile("/proc/meminfo")
    if err != nil {
        return ""
    }
   lines := strings.Split(string(contents), "\n")
   line := lines[0];
   dato := strings.Replace(string(line), "MemTotal:", "", 1)
   Total := strings.Replace(string(dato), " ", "", 10)
   Total2 := strings.Replace(string(Total), "kB", "", 1)
   fmt.Println("Total de RAM " + Total2);

   line2 := lines[1];
   dato2 := strings.Replace(string(line2), "MemFree:", "", 1)
   Libre := strings.Replace(string(dato2), " ", "", 15)
   Libre2 := strings.Replace(string(Libre), "kB", "", 1)
   fmt.Println("RAM Libre " + Libre2);

   i1, err := strconv.Atoi(Total2)
   i2, err := strconv.Atoi(Libre2)
   usado := i1 - i2;
   porcentaje := usado * 100 / i1;
    fmt.Println("RAM usada", porcentaje, "%");
    
    s := fmt.Sprintf("%d", porcentaje)
    s = s + " " + Total2
    fmt.Println("soy el CAN " + s);
   return s;
}

func ramPage(w http.ResponseWriter, r *http.Request) {
    html := `<!DOCTYPE html>

<html>
   <head>
        <meta charset="UTF-8">
        <link rel="stylesheet" href="/static/stylesheets/main.css">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">

        <!--stylesheet PLATO-->
    <meta name="description" content="Clean responsive bootstrap website template">
    <meta name="author" content="">
    <!-- styles -->
    <link href="/static/stylesheets/Plato/assets/css/bootstrap.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/bootstrap-responsive.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/docs.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/prettyPhoto.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/js/google-code-prettify/prettify.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/flexslider.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/refineslide.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/font-awesome.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/animate.css" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400italic,400,600,700" rel="stylesheet">

    <link href="/static/stylesheets/Plato/assets/css/style.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/color/default.css" rel="stylesheet">

    <!-- fav and touch icons -->
    <link rel="shortcut icon" href="assets/ico/favicon.ico">
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="assets/ico/apple-touch-icon-144-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="114x114" href="assets/ico/apple-touch-icon-114-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="72x72" href="assets/ico/apple-touch-icon-72-precomposed.png">
    <link rel="apple-touch-icon-precomposed" href="assets/ico/apple-touch-icon-57-precomposed.png">
    <script src='http://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js'></script>
    <script src="js/vendor/modernizr-2.8.3.min.js"></script>
    <script src="js/jquery.flot.js"  type="text/javascript"></script>
    <script src="js/jquery-1.12.0.min.js"  type="text/javascript"></script>
    <script src="https://canvasjs.com/assets/script/canvasjs.min.js"> </script>

    <!--stylesheet PLATO-->
        <!-- The welcome struct (shown in the main.go code) is received within the HTML and we just need to use the . operator and retrieve the information we want -->
        <title>RAM</title>
   </head>







   <body>
  <header>
    <!-- Navbar
    ================================================== -->
    <div class="cbp-af-header">
      <div class=" cbp-af-inner">
        <div class="container">
          <div class="row">

            <div class="span4">
              <!-- logo -->
              <div class="logo">
                <h1><a href="index.html">SOPES 1</a></h1>
                <!-- <img src="assets/img/logo.png" alt="" /> -->
              </div>
              <!-- end logo -->
            </div>

            <!-- top menu -->
            <div class="navbar">
              <div class="navbar-inner">
                <nav>
                  
                </nav>
              </div>
            </div>
            <!-- end menu -->
            
            </div>

          </div>
        </div>
      </div>
    </div>
  </header>
  <section id="intro">

    <div class="container">
      <div class="row">
        <div class="span6">
          <h2><strong>RAM<span class="highlight primary">Monitoring</span></strong></h2>
          <p class="lead">
          </p>



            <script>
              // Attach a submit handler to the form
              $( "#cpuOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#ramOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#processOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#logoutOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>
              <!-- end menu -->


        </div>
        <div class="span6">
      </div>
    </div>
   <div id='result1'><h3></h3></div><br><br>
   <div id='result2'><h3></h3></div><br><br>
   <div id='result3'><h3></h3></div><br><br>
   <div id='result4'><h3></h3></div><br><br>
   <div id='result5'><h3></h3></div><br><br>
   <div id='result6'><h3></h3></div><br><br>
   <div id='result40'><h3></h3></div><br><br>
   <div id='result60'><h3></h3></div><br><br>
   <div id='result65'><h3></h3></div><br><br>
    <script>
     $(function() { // Ojo! uso jQuery, recuerda añadirla al html
      var porcent = 0;
      var libre = 0;
      var ram1 = 0;
      var ram2 = 0;
      var ram3 = 0;
      var ram4 = 0;
      var ram5 = 0;
      cron(); // Lanzo cron la primera vez
      function cron() {
          $.ajax({
              url: 'receive',
              type: 'post',
              dataType: 'html',
              data : { ajax_post_data: 'hello'},
              success : function(data) {

                ram1 = ram2;
                ram2 = ram3;
                ram3 = ram4;
                ram4 = ram5;
                ram5 = parseInt(data.split(" ")[0],10);

                $('#result40').html("Porcentaje RAM Usado: " + (data.split(" ")[0]));
                $('#result60').html("Total RAM: " + (data.split(" ")[1] / 1024) + " MB");
                $('#result65').html("Total RAM consumida: " + ((data.split(" ")[1]) * (data.split(" ")[0]) / 1024 / 100) + " MB");
                              
                  var chart2 = new CanvasJS.Chart("result1", {
                      animationEnabled: true,
                      width: 600,
                      height: 300,
                      theme: "light2",
                      title:{
                          text: "Ram Utilizado"
                      },
                      data: [{        
                          type: "line",
                          dataPoints: [
                              { y: ram1 },
                              { y: ram2 },
                              { y: ram3 },
                              { y: ram4 },
                              { y: ram5 }
                          ]
                      }]
                  });
                  chart2.render();
              },
            });
      }
      setInterval(function() {
          
          cron();
      }, 3000); // Lanzará la petición cada 10 segundos
  });
</script>
  </section>

  <!-- Footer
 ================================================== -->
  <footer class="footer">
    <div class="container">
      <div class="row">
        <div class="span3">
          <div class="widget">
            <!-- logo -->
            <div class="footerlogo">
              <h6><a href="index.html">Plato</a></h6>
              <!-- <img src="assets/img/logo.png" alt="" /> -->
            </div>
            <!-- end logo -->
            <address>
        <strong>USAC</strong><br>
        Sistemas Operativos 1 "A"<br>
        Segundo Semestre 2019<br>
          </div>
        </div>
        <div class="span3">
          <div class="widget">
            <h5>SISTEMA DE MONITOREO</h5>
            <div class="flickr_badge">
              <img src="/static/stylesheets/ubuntu.svg">
            </div>
            <div class="clear"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="subfooter">
      <div class="container">
        <div class="row">
          <div class="span6">
            <p>
              &copy; Plato - All right reserved
            </p>
          </div>
          <div class="span6">
            <div class="pull-right">
              <div class="credits">
                <!--
                  All the links in the footer should remain intact.
                  You can delete the links only if you purchased the pro version.
                  Licensing information: https://bootstrapmade.com/license/
                  Purchase the pro version with working PHP/AJAX contact form: https://bootstrapmade.com/buy/?theme=Plato
                -->
                Designed by <a href="https://bootstrapmade.com/">BootstrapMade</a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </footer>

  <script src="/static/stylesheets/Plato/assets/js/jquery.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/modernizr.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.easing.1.3.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/google-code-prettify/prettify.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/bootstrap.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.prettyPhoto.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/portfolio/jquery.quicksand.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/portfolio/setting.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/hover/jquery-hover-effect.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.flexslider.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/classie.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/cbpAnimatedHeader.min.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.refineslide.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.ui.totop.js"></script>

  <!-- Template Custom Javascript File -->
  <script src="/static/stylesheets/Plato/assets/js/custom.js"></script>
  <script src="https://code.jquery.com/jquery-3.4.1.js"></script>
</body>

</html>`

    w.Write([]byte(fmt.Sprintf(html)))
    ram.ramGraph = html;
}


func cpuPage(w http.ResponseWriter, r *http.Request) {
    html := `<!DOCTYPE html>

<html>
   <head>
        <meta charset="UTF-8">
        <link rel="stylesheet" href="/static/stylesheets/main.css">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">

        <!--stylesheet PLATO-->
    <meta name="description" content="Clean responsive bootstrap website template">
    <meta name="author" content="">
    <!-- styles -->
    <link href="/static/stylesheets/Plato/assets/css/bootstrap.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/bootstrap-responsive.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/docs.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/prettyPhoto.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/js/google-code-prettify/prettify.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/flexslider.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/refineslide.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/font-awesome.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/css/animate.css" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400italic,400,600,700" rel="stylesheet">

    <link href="/static/stylesheets/Plato/assets/css/style.css" rel="stylesheet">
    <link href="/static/stylesheets/Plato/assets/color/default.css" rel="stylesheet">

    <!-- fav and touch icons -->
    <link rel="shortcut icon" href="assets/ico/favicon.ico">
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="assets/ico/apple-touch-icon-144-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="114x114" href="assets/ico/apple-touch-icon-114-precomposed.png">
    <link rel="apple-touch-icon-precomposed" sizes="72x72" href="assets/ico/apple-touch-icon-72-precomposed.png">
    <link rel="apple-touch-icon-precomposed" href="assets/ico/apple-touch-icon-57-precomposed.png">
    <script src='http://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js'></script>
    <script src="js/vendor/modernizr-2.8.3.min.js"></script>
    <script src="js/jquery.flot.js"  type="text/javascript"></script>
    <script src="js/jquery-1.12.0.min.js"  type="text/javascript"></script>
    <script src="https://canvasjs.com/assets/script/canvasjs.min.js"> </script>

    <!--stylesheet PLATO-->
        <!-- The welcome struct (shown in the main.go code) is received within the HTML and we just need to use the . operator and retrieve the information we want -->
        <title>CPU</title>
   </head>







   <body>
  <header>
    <!-- Navbar
    ================================================== -->
    <div class="cbp-af-header">
      <div class=" cbp-af-inner">
        <div class="container">
          <div class="row">

            <div class="span4">
              <!-- logo -->
              <div class="logo">
                <h1><a href="index.html">SOPES 1</a></h1>
                <!-- <img src="assets/img/logo.png" alt="" /> -->
              </div>
              <!-- end logo -->
            </div>

            <!-- top menu -->
            <div class="navbar">
              <div class="navbar-inner">
                <nav>

                </nav>
              </div>
            </div>
            <!-- end menu -->
            
            </div>

          </div>
        </div>
      </div>
    </div>
  </header>
  <section id="intro">

    <div class="container">
      <div class="row">
        <div class="span6">
          <h2><strong>CPU<span class="highlight primary">Monitoring</span></strong></h2>
          <p class="lead">
          </p>



              <script>
              // Attach a submit handler to the form
              $( "#cpuOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#ramOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#processOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>

              <script>
              // Attach a submit handler to the form
              $( "#logoutOption" ).submit(function( event ) {
               
                // Stop form from submitting normally
                event.preventDefault();
               
                // Get some values from elements on the page:
                var $form = $( this ),
                  term = $form.find( "input[name='s']" ).val(),
                  url = $form.attr( "action" );
               
                // Send the data using post
                var posting = $.post( url, { s: term } );
               
                // Put the results in a div
                posting.done(function( data ) {
                  var content = $( data ).find( "#content" );
                  $( "#result" ).empty().append( content );
                });
              });
              </script>
              <!-- end menu -->

        </div>
        <div class="span6">
      </div>
    </div>
   
   <div id='result7'><h3>.</h3></div><br><br>
   <div id='result72'><h3>.</h3></div><br><br>
   <div id='result73'><h3>.</h3></div><br><br>
   <div id='result74'><h3>.</h3></div><br><br>
   <div id='result75'><h3>..</h3></div><br><br>

    <script>
     $(function() { // Ojo! uso jQuery, recuerda añadirla al html
      var d1 = 0;
      var d2 = 0;
      var d3 = 0;
      var d4 = 0;
      var d5 = 0;
      var d6 = 0;
      var d7 = 0;
      var d8 = 0;
      var d9 = 0;
      var d10 = 0;
      var porcent = 0;
      var libre = 0;
      cron(); // Lanzo cron la primera vez
      function cron() {
          $.ajax({
              url: 'receive2',
              type: 'post',
              dataType: 'html',
              data : { ajax_post_data2: 'hello'},
              success : function(data) {
                d1 = d2;
                d2 = d3;
                d3 = d4;
                d4 = d5;
                d5 = d6;
                d6 = d7;
                d7 = d8;
                d8 = d9;
                d9 = d10;
                d10 = parseFloat(data.split(" ")[0],10);

                $('#result75').html("Porcentaje CPU Actual: " + data.split(" ")[0]);
                
                  var chart = new CanvasJS.Chart("result7", {
                      animationEnabled: true,
                      width: 600,
                      height: 300,
                      theme: "light2",
                      title:{
                          text: "CPU Utilizado"
                      },
                      data: [{        
                          type: "line",
                          dataPoints: [
                              { y: d1 },
                              { y: d2 },
                              { y: d3 },
                              { y: d4 },
                              { y: d5 },
                              { y: d6 },
                              { y: d7 },
                              { y: d8 },
                              { y: d9 },
                              { y: d10 }
                          ]
                      }]
                  });
                  chart.render();
              },
            });
      }
      setInterval(function() {
          
          cron();
      }, 3000); // Lanzará la petición cada 10 segundos
  });
</script>


  </section>

  <!-- Footer
 ================================================== -->
  <footer class="footer">
    <div class="container">
      <div class="row">
        <div class="span3">
          <div class="widget">
            <!-- logo -->
            <div class="footerlogo">
              <h6><a href="index.html">Plato</a></h6>
              <!-- <img src="assets/img/logo.png" alt="" /> -->
            </div>
            <!-- end logo -->
            <address>
        <strong>USAC</strong><br>
        Sistemas Operativos 1 "A"<br>
        Segundo Semestre 2019<br>
          </div>
        </div>
        <div class="span3">
          <div class="widget">
            <h5>SISTEMA DE MONITOREO</h5>
            <div class="flickr_badge">
              <img src="/static/stylesheets/ubuntu.svg">
            </div>
            <div class="clear"></div>
          </div>
        </div>
      </div>
    </div>
    <div class="subfooter">
      <div class="container">
        <div class="row">
          <div class="span6">
            <p>
              &copy; Plato - All right reserved
            </p>
          </div>
          <div class="span6">
            <div class="pull-right">
              <div class="credits">
                <!--
                  All the links in the footer should remain intact.
                  You can delete the links only if you purchased the pro version.
                  Licensing information: https://bootstrapmade.com/license/
                  Purchase the pro version with working PHP/AJAX contact form: https://bootstrapmade.com/buy/?theme=Plato
                -->
                Designed by <a href="https://bootstrapmade.com/">BootstrapMade</a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </footer>

  <script src="/static/stylesheets/Plato/assets/js/jquery.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/modernizr.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.easing.1.3.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/google-code-prettify/prettify.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/bootstrap.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.prettyPhoto.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/portfolio/jquery.quicksand.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/portfolio/setting.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/hover/jquery-hover-effect.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.flexslider.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/classie.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/cbpAnimatedHeader.min.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.refineslide.js"></script>
  <script src="/static/stylesheets/Plato/assets/js/jquery.ui.totop.js"></script>

  <!-- Template Custom Javascript File -->
  <script src="/static/stylesheets/Plato/assets/js/custom.js"></script>
  <script src="https://code.jquery.com/jquery-3.4.1.js"></script>
</body>

</html>`

    w.Write([]byte(fmt.Sprintf(html)))
    ram.ramGraph = html;
}

func ramGraphHTML () string{
  return "puroiwejroiwjeriowejior"
}

func receiveAjax(w http.ResponseWriter, r *http.Request) {
   if r.Method == "POST" {
        ajax_post_data := r.FormValue("ajax_post_data")
        fmt.Println("Receive ajax post data string ", ajax_post_data)
        fmt.Println();
        w.Write([]byte(Calculos()))
        //w.Write([]byte(CalculosCPU()))
   }
}

func receiveAjax2(w http.ResponseWriter, r *http.Request) {
   if r.Method == "POST" {
        ajax_post_data2 := r.FormValue("ajax_post_data2")
        fmt.Println("Receive ajax post data string ", ajax_post_data2)
        fmt.Println();
        w.Write([]byte(CalculosCPU()))
   }
}

func Calculos() string{
    s := getRAMSample();
    return s;
}

func CalculosCPU() string{
  return fmt.Sprintf("%f", percentCPU())
}

func Corrimiento(texto string){
    cpus[0] = cpus[1];
    cpus[1] = cpus[2];
    cpus[2] = cpus[3];
    cpus[3] = cpus[4];
    cpus[4] = texto;
}

func percentCPU() float64{
    cmd := exec.Command("ps", "aux")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    processes := make([]*Process, 0)
    for {
        line, err := out.ReadString('\n')
        if err!=nil {
            break;
        }
        tokens := strings.Split(line, " ")
        ft := make([]string, 0)
        for _, t := range(tokens) {
            if t!="" && t!="\t" {
                ft = append(ft, t)
            }
        }
        //log.Println(len(ft), ft)
        pid, err := strconv.Atoi(ft[1])
        if err!=nil {
            continue
        }
        cpu, err := strconv.ParseFloat(ft[2], 64)
        if err!=nil {
            log.Fatal(err)
        }
        processes = append(processes, &Process{pid, cpu})
    }
    var percent float64
    percent = 0
    for _, p := range(processes) {
        //fmt.Println("Process ", p.pid, " takes ", p.cpu, " % of the CPU")
        percent = percent + p.cpu
    }
    fmt.Println("Porcentaje CPU ", percent)
    return percent
}

func countProcess() string{
    out, err := exec.Command("/bin/sh", "-c", "ps -A --no-headers | wc -l").Output()
    if err != nil {
        log.Fatal(err)
    }
    var count string
    count = string(out)
    fmt.Printf("Number of running processes: %s\n", count)
    return count
}

func countProcessURL(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(countProcess)
}