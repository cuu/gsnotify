package main

import (
	"os"
	"fmt"
	"os/exec"
	"io/ioutil"
	"strings"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"path/filepath"
	"net"
	gotime "time"
  
	"github.com/go-ini/ini"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	
	"github.com/cuu/gogame"
	"github.com/cuu/gogame/color"
	"github.com/cuu/gogame/display"
	"github.com/cuu/gogame/surface"
	"github.com/cuu/gogame/event"
	//"github.com/cuu/gogame/rect"
	"github.com/cuu/gogame/draw"
	//"github.com/cuu/gogame/image"
	"github.com/cuu/gogame/font"
	"github.com/cuu/gogame/time"	
  
	"github.com/clockworkpi/LauncherGoDev/sysgo"
	"github.com/clockworkpi/LauncherGoDev/sysgo/easings"
	"github.com/clockworkpi/LauncherGoDev/sysgo/UI"
	
)

type SDLWindow struct{
	PosX int
	PosY int
	OnScreen bool  // default false== hide
	Data100 []int
	Data40  []int
	win *sdl.Window 
	screen *sdl.Surface
	main_font *ttf.Font
	
    
}

var sdl_window *SDLWindow

type JobRespond struct{
    Type string  `json:"type"`    //once,repeat
    Content string   `json:"content"`
}

func (self *JobRespond) String() string{
    return self.Type+"-"+self.Content
}

var (
	JobMap map[string]string // BashName => String(JobRespond)
	ALLOW_EXTS=[5]string{".sh",".py",".lsp",".js",".bin"}
	Width = 320
	Height = 24
	SKIP_READ_DIR = 2
	DELAY_MS = 2000
	DELAY_FREQ = 30*1000   
	BGCOLOR  =  &color.Color{0xff,0x00,0x4d,255}
	TXTCOLOR =  &color.Color{0xff,0xff,0xff,255}
	FTSIZE = 14
	Enabled = true
	AutoShutDown = true
)

const (
	RUNEVT = 1 
	EasingDur = 10
	EasingDelay = 20 //ms
	GSNOTIFY_CFG="gsnotify.cfg"
)

var done chan bool

var Cfg *ini.File

func ConvertToRGB(hexstr string) *color.Color {
	if len(hexstr) < 7 || string(hexstr[0]) != "#" { // # 00 00 00
		log.Fatalf("ConvertToRGB hex string format error %s", hexstr)
		//fmt.Printf("ConvertToRGB hex string format error %s", hexstr)
		return nil
	}
	
	h := strings.TrimLeft(hexstr,"#")

	r,_ := strconv.ParseInt(h[0:2], 16,0)
	g,_ := strconv.ParseInt(h[2:4], 16,0)
	b,_ := strconv.ParseInt(h[4:6], 16,0)
	
	col := &color.Color{ uint32(b),uint32(g),uint32(r),255 }
	return col
}


func DumpConfig() {
	if UI.FileExists(GSNOTIFY_CFG)== false {
		sample_ini :=  fmt.Sprintf("[Settings]\nDELAY_MS=%d\nDELAY_FREQ=%d\nBGCOLOR=#ff004d\nTXTCOLOR=#ffffff\nWidth=%d\nHeight=%d\nFTSIZE=%d\nEnabled=%s\nAutoShutDown=%s",DELAY_MS,DELAY_FREQ,Width,Height,FTSIZE, strings.Title(fmt.Sprintf("%t",Enabled)),strings.Title(fmt.Sprintf("%t",AutoShutDown) ) )

		f, err := os.Create(GSNOTIFY_CFG)
		UI.Assert(err)
		n, err := f.WriteString(sample_ini)
    fmt.Printf("wrote %d bytes\n", n)
		f.Sync()
		f.Close()
	}
}

func WriteConfig() {
	
	if UI.FileExists(GSNOTIFY_CFG) {

		if Cfg != nil {
			section := Cfg.Section("Settings")

			if section.HasKey("Enabled") == false {
				section.NewKey("Enabled","True")
			}
			
			if Enabled == true {
				section.Key("Enabled").SetValue("True")
			}
			if Enabled == false {
				section.Key("Enabled").SetValue("False")
			}
		
			if AutoShutDown == true {
				section.Key("AutoShutDown").SetValue("True")
			}
			if Enabled == false {
				section.Key("AutoShutDown").SetValue("False")
			}
	
			err:= Cfg.SaveTo(GSNOTIFY_CFG)
			if err != nil {
				fmt.Println(err)
			}
			
		}
		
	}else {
		DumpConfig()
		LoadConfig()
	}
}


func LoadConfig() {
	if UI.FileExists(GSNOTIFY_CFG) == false {
		DumpConfig()
	}
	
	if UI.FileExists(GSNOTIFY_CFG) {
		load_opts := ini.LoadOptions{
			IgnoreInlineComment:true,
		}
		var err error
		
		Cfg, err = ini.LoadSources(load_opts, GSNOTIFY_CFG )
		if err != nil {
			fmt.Printf("Fail to read file: %v\n", err)
			return
		}
		
		section := Cfg.Section("Settings")
		if section != nil {
			_opts := section.KeyStrings()
			for _,v := range _opts {
				if v == "DELAY_MS" {
					val := section.Key(v).String()
					i, err := strconv.Atoi(val)
					if err == nil {
						DELAY_MS = i
					}
				}
				if v == "DELAY_FREQ" {
					val := section.Key(v).String()
					i, err := strconv.Atoi(val)
					if err == nil {
						DELAY_FREQ = i
					}
				}
				
				if v == "Width" {
					val := section.Key(v).String()
					i, err := strconv.Atoi(val)
					if err == nil {
						Width = i
					}
				}
				
				if v == "Height" {
					val := section.Key(v).String()
					i, err := strconv.Atoi(val)
					if err == nil {
						Height = i
					}
				}
				
				if v == "FTSIZE" {
					val := section.Key(v).String()
					i, err := strconv.Atoi(val)
					if err == nil {
						FTSIZE = i
					}
				}
				
				if v == "BGCOLOR" {
					parsed_color := ConvertToRGB( section.Key(v).String() )
					if parsed_color != nil {
						BGCOLOR = parsed_color
					}
				}
				
				if v == "TXTCOLOR" {
					parsed_color := ConvertToRGB( section.Key(v).String() )
					if parsed_color != nil {
						TXTCOLOR = parsed_color
					}
				}
				
				if v == "Enabled" {
					if section.Key(v).String() == "True" {
						Enabled = true
					}

					if section.Key(v).String() == "False" {
						Enabled = false
					}
				}

				if v == "AutoShutDown" {
					if section.Key(v).String() == "True" {
					 	AutoShutDown = true
					}

					if section.Key(v).String() == "False" {
						AutoShutDown = false
					}
				}
			
			}
		}
	}
}

func TheInit() {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	os.Chdir(dir)

	Enabled  = true

	LoadConfig()
	
	
	sdl_window = &SDLWindow{}
	sdl_window.Data100 = EasingData(0,100)
	sdl_window.Data40 = EasingData(0,40)
	
	JobMap = make(map[string]string)
	
	done = make(chan bool)
	
}

func EasingData(start,distance int) []int {
	current_time := float32(0.0)
	start_posx   := float32(0.0)
	current_posx := start_posx
	final_posx   := float32(distance)
//	posx_init    := start
	dur          := EasingDur
	last_posx    := float32(0.0)

	var all_last_posx []int

	for i:=0;i<distance*dur;i++ {
		current_posx = float32(easings.SineIn(float32(current_time), float32(start_posx), float32(final_posx-start_posx),float32(dur)))
		if current_posx >= final_posx {
			current_posx = final_posx
		}
		dx := current_posx - last_posx
		all_last_posx = append(all_last_posx,int(dx))
		current_time+=1.0
		last_posx = current_posx
		if current_posx >= final_posx {
			break
		}
	}

	c := 0
	for _,v := range all_last_posx {
		c+=v
	}
	if c < int(final_posx - start_posx) {
		all_last_posx = append(all_last_posx, int( int(final_posx) - c ))
	}

	return all_last_posx	
}

func (self *SDLWindow) UpdateWindowPos() {
    x,y := display.GetWindowPos(self.win)
    self.PosX = x
    self.PosY = y
}

func (self *SDLWindow) EasingWindowRight(distance int) {
    data := EasingData(0, distance)
    for _,v := range  data  {
        self.PosX += v
        display.SetWindowPos(self.win,self.PosX,self.PosY)
        time.BlockDelay(EasingDelay)
    }
}

func (self *SDLWindow) EasingWindowLeft(distance int) {
    data := EasingData(0,distance)
    for _,v := range data {
        self.PosX -= v
        display.SetWindowPos(self.win,self.PosX,self.PosY)
        time.BlockDelay(EasingDelay)
    }
}

func (self *SDLWindow) EasingWindowTop(distance int) {
    data := EasingData(0,distance)
    for _,v := range data {
        self.PosY -= v
        display.SetWindowPos(self.win,self.PosX,self.PosY)
        time.BlockDelay(EasingDelay)
        display.Flip()
    }
}

func (self *SDLWindow) EasingWindowBottom(distance int) {
    data := EasingData(0,distance)
    for _,v := range data {
        self.PosY += v
        display.SetWindowPos(self.win,self.PosX,self.PosY)
        time.BlockDelay(EasingDelay)
        display.Flip()
    }
}


func CheckScriptExt(_script_name string ) bool {
    
    for _,v := range ALLOW_EXTS {

        if strings.HasSuffix(strings.ToLower(_script_name), v) {
            return true
        }
    }
    
    return false
}

func RunScript(_script_name string) *JobRespond {
    cur_time_unix := time.Unix()

    if UI.IsAFile(_script_name) {
        if CheckScriptExt(_script_name) {
            out, err := exec.Command(_script_name,fmt.Sprintf("%d",cur_time_unix)).Output()
            if err != nil {
                fmt.Println(err)
                
            }else{
                if len(out) < 13 {
                    return nil
                }
                //fmt.Println( fmt.Sprintf("%d",cur_time_unix)  )
                jr := &JobRespond{}
                err = json.Unmarshal(out, jr)
                if err == nil {
                    return jr
                }else{
                    fmt.Println(err)
                }
            }
        }
    }
    
    return nil
}

func ShowARound(content string) {
    
	surface.Fill(sdl_window.screen, BGCOLOR) 
    
 	my_text := font.Render(sdl_window.main_font,content,true, TXTCOLOR,nil)

	surface.Blit(sdl_window.screen,my_text,
		draw.MidRect(Width/2,Height/2-2,surface.GetWidth(my_text),surface.GetHeight(my_text),Width,Height),nil)
	
	display.Flip()    
	
	sdl_window.EasingWindowBottom(Height)
	
	time.BlockDelay(DELAY_MS/2)
	
	sdl_window.EasingWindowTop(Height)
    
}

func LoopCheckJobs(_dir string) {
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		if (recover() != nil) {
			err := errors.New("execute script errors")
			fmt.Println(err)
		}
	}()
	
	counter := 0
	
	if UI.FileExists(_dir) == false && UI.IsDirectory(_dir) == false {
		return
	}
	
	os.Chdir(_dir)
	_dir = "."
	var files []os.FileInfo
	var err error 
	for {
    
		if counter == 0 {
			files,err = ioutil.ReadDir(_dir)
			if err != nil {
				log.Fatal(err)
				return
			}
		}
		
		if Enabled == true {
			for _,f := range files {
			
				fname := _dir+"/"+f.Name()
				if CheckScriptExt(fname) == false {
					continue
				}
				
				job_respond := RunScript( fname )
				if job_respond!= nil {
					job_respond_string := job_respond.String()
					
					if job_respond.Type == "once" {
						if val, ok := JobMap[_dir+"/"+f.Name()]; ok {
							
							if val != job_respond_string {
								JobMap[_dir+"/"+f.Name()] = job_respond_string
								ShowARound( job_respond.Content )
							}
						}else {
							JobMap[_dir+"/"+f.Name()] = job_respond_string
							ShowARound( job_respond.Content )
						}
					}else if job_respond.Type == "repeat" {
						ShowARound( job_respond.Content )
					}
				}
			}
		} else {
			
		}
		//fmt.Println("DELAY_FREQ:",DELAY_FREQ)	
		time.BlockDelay(DELAY_FREQ) 
		
		counter+=1
		if counter >= SKIP_READ_DIR {
			counter = 0
		}
	}
	
}

func  GetBatteryPercent() int {
  if UI.FileExists(sysgo.Battery) == false {
    return -1
  }
  
  batinfos,err := UI.ReadLines(sysgo.Battery)
  if err == nil {
    for _,v := range batinfos {
      if strings.HasPrefix(v, "POWER_SUPPLY_STATUS") {
        parts := strings.Split(v,"=")
        if len(parts) > 1 {
          if strings.Contains(parts[1],"Charging") {
            return 100//ignore Charging status
          }
        }
      }else if strings.HasPrefix(v,"POWER_SUPPLY_CAPACITY") {
        parts := strings.Split(v,"=")
        if len(parts) > 1 {
          cur_cap,err := strconv.Atoi(parts[1])
          if err == nil {
            return cur_cap
          }else {
            return 100 
          }
        }
      }
    }
  }else{
    fmt.Println(err)
  }
  //any error happens,return 100
  return 100  
}

func ShutDownWhenLowPower() {
  for {
    bat := GetBatteryPercent()
  
    if bat > -1 {
    
      if bat >= 0 && bat < 2 {// last 10 % of battery is so strong
      
        cmd := exec.Command("sudo","halt","-p")
        cmd.Run()
        os.Exit(0)
      
      }
    
    }else {
      fmt.Println("Battery not existed")
    }
  
    gotime.Sleep(gotime.Duration(10) * gotime.Second)
  }
}

func run() int {

	display.Init()
	mode,_ := display.GetCurrentMode(0)
  
	sdl_window.screen = display.SetMode(int32(Width),int32(Height), 0 ,32)
	sdl_window.win = display.GetWindow()
  
  
	display.SetWindowPos(sdl_window.win,(int(mode.W)-Width)/2, -Height)
	
	display.SetWindowTitle(sdl_window.win,"GameShellNotify")
	display.SetWindowBordered(sdl_window.win,false)
	
	surface.Fill(sdl_window.screen, &color.Color{255,255,255,255} ) 
	
	fmt.Println(sdl_window.screen.Pitch)
	fmt.Println(sdl_window.screen.BytesPerPixel() )
	
	font.Init()
	
	font_path := "/home/cpi/apps/launcher/skin/default/truetype/NotoSansCJK-Regular.ttf"
	if UI.FileExists(font_path) == false {
		font_path = "/home/cpi/launcher/skin/default/truetype/NotoSansCJK-Regular.ttf"
	}
	notocjk := font.Font(font_path,FTSIZE)
	fmt.Println( font.LineSize( notocjk ))
	sdl_window.main_font = notocjk
	
 	my_text := font.Render(notocjk,"GsNotify",true, TXTCOLOR , nil)
	
	surface.Blit(sdl_window.screen,my_text,
		draw.MidRect(Width/2,Height/2,surface.GetWidth(my_text),surface.GetHeight(my_text),Width,Height),nil)
	
	display.Flip()
	
	event.AddCustomEvent(RUNEVT)
	
	sdl_window.UpdateWindowPos()
	
	go LoopCheckJobs("Jobs")
	
	go ShutDownWhenLowPower()
  
	syswminfo,_ := sdl_window.win.GetWMInfo()
	x11info := syswminfo.GetX11Info()
	fmt.Println("x11info: ",x11info.Window )
	
	running := true
	
	
	for running {
		ev := event.Wait()
		if ev.Type == event.QUIT {
			running = false
			break
		}
		if ev.Type == event.USEREVENT {
			
			fmt.Println("UserEvent:"+ev.Data["Msg"])
		}
		if ev.Type == event.KEYDOWN {
			fmt.Println(ev)
			if ev.Data["Key"] == "Q" {
				gogame.Quit()
				return 0
			}
			if ev.Data["Key"] == "Escape" {
				return 0
			}
			if ev.Data["Key"] == "L" {
				sdl_window.EasingWindowLeft(100)
			}
			if ev.Data["Key"] == "R" {
				sdl_window.EasingWindowRight(100)
			} 
			if ev.Data["Key"] == "T" {
				sdl_window.EasingWindowTop(40)
			}                        
			if ev.Data["Key"] == "P" {				
				event.Post(RUNEVT,"GODEBUG=cgocheck=0 sucks") // just id and string, simplify the stuff
			}
		}
	}
	
	return 0
}

func SearchAndDestory(pidFile string) {
    
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				process.Kill()
				ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
				return
			}
		}
	}
	
	ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	return
}

func HandleCMD(c net.Conn) {
	
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
		println("Got:", string(data))
		data_str := strings.TrimRight(string(data),"\r\n")
		
		if data_str == "Enable" {
			fmt.Println("Enable gsnotify")
			Enabled = true
			WriteConfig()
		}
		
		if data_str == "Disable" {
			fmt.Println("Disable gsnotify now")
			Enabled = false
			WriteConfig()
		}
		
		_, err = c.Write(data)		
		if err != nil {
			log.Fatal("Writing client error: ", err)
		}
	}	
}

func InitSocket() {
	cmdName := "rm"
	cmdArgs := []string{"-rf","/tmp/gsnotify.sock"}

	if _, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "Error binding unix socket to /tmp/gsnotify.sock ", err)
		return
	}
	
	ln, err := net.Listen("unix", "/tmp/gsnotify.sock")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}
		
		go HandleCMD(fd)
	}
}



func main() {
	var exitcode int
  
	os.Setenv("GODEBUG", "cgocheck=0")

  if len(os.Args) == 1 {
    TheInit()
    go InitSocket()	
    SearchAndDestory("/tmp/gsnotify.pid")
  
    sdl.Main(func() {
      time.BlockDelay(DELAY_FREQ/2)
      exitcode = run()
    })
    os.Exit(exitcode)
  }else { // if has any arguments,turn into daemon only for battery lower than 3% to poweroff
    ShutDownWhenLowPower()
  }
}
