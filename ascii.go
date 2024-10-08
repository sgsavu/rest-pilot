package main

import (
	"math/rand"
	"time"
)

var asciiArtPresets = []string{
	`
██████╗ ███████╗███████╗████████╗    ██████╗ ██╗██╗     ██████╗ ████████╗
██╔══██╗██╔════╝██╔════╝╚══██╔══╝    ██╔══██╗██║██║     ██╔══██╗╚══██╔══╝
██████╔╝█████╗  ███████╗   ██║       ██████╔╝██║██║     ██║  ██║   ██║   
██╔══██╗██╔══╝  ╚════██║   ██║       ██╔═══╝ ██║██║     ██║  ██║   ██║   
██║  ██║███████╗███████║   ██║       ██║     ██║███████╗██████╔╝   ██║   
╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝       ╚═╝     ╚═╝╚══════╝╚═════╝    ╚═╝   
   `,
	`
                                     o                   o     o                 o     
                                    <|>                _<|>_  <|>               <|>    
                                    < >                       / \               < >    
 \o__ __o     o__  __o       __o__   |      \o_ __o      o    \o/    o__ __o     |     
  |     |>   /v      |>     />  \    o__/_   |    v\    <|>    |    /v     v\    o__/_ 
 / \   < >  />      //      \o       |      / \    <\   / \   / \  />       <\   |     
 \o/        \o    o/         v\      |      \o/     /   \o/   \o/  \         /   |     
  |          v\  /v __o       <\     o       |     o     |     |    o       o    o     
 / \          <\/> __/>  _\o__</     <\__   / \ __/>    / \   / \   <\__ __/>    <\__  
                                            \o/                                        
                                             |                                         
                                            / \                                        
`,
	`
               _         _ _       _   
              | |       (_) |     | |  
 _ __ ___  ___| |_ _ __  _| | ___ | |_ 
| '__/ _ \/ __| __| '_ \| | |/ _ \| __|
| | |  __/\__ \ |_| |_) | | | (_) | |_ 
|_|  \___||___/\__| .__/|_|_|\___/ \__|
                  | |                  
                  |_|                  
`,
	`                                                                              
                                                       ###                         
                                                   #    ###                        
                                    #             ###    ##                  #     
                                   ##              #     ##                 ##     
                                   ##                    ##                 ##     
###  /###     /##       /###     ######## /###   ###     ##      /###     ######## 
 ###/ #### / / ###     / #### / ######## / ###  / ###    ##     / ###  / ########  
  ##   ###/ /   ###   ##  ###/     ##   /   ###/   ##    ##    /   ###/     ##     
  ##       ##    ### ####          ##  ##    ##    ##    ##   ##    ##      ##     
  ##       ########    ###         ##  ##    ##    ##    ##   ##    ##      ##     
  ##       #######       ###       ##  ##    ##    ##    ##   ##    ##      ##     
  ##       ##              ###     ##  ##    ##    ##    ##   ##    ##      ##     
  ##       ####    /  /###  ##     ##  ##    ##    ##    ##   ##    ##      ##     
  ###       ######/  / #### /      ##  #######     ### / ### / ######       ##     
   ###       #####      ###/        ## ######       ##/   ##/   ####         ##    
                                       ##                                          
                                       ##                                          
                                       ##                                          
                                        ##                                         
`,
	`                                         
                    ,           ,,         ,  
                   ||         ' ||        ||  
,._-_  _-_   _-_, =||= -_-_  \\ ||  /'\\ =||= 
 ||   || \\ ||_.   ||  || \\ || || || ||  ||  
 ||   ||/    ~ ||  ||  || || || || || ||  ||  
 \\,  \\,/  ,-_-   \\, ||-'  \\ \\ \\,/   \\, 
                       |/                     
                       '                      
`,
	`
                         __         .__.__          __   
_______   ____   _______/  |_______ |__|  |   _____/  |_ 
\_  __ \_/ __ \ /  ___/\   __\____ \|  |  |  /  _ \   __\
 |  | \/\  ___/ \___ \  |  | |  |_> >  |  |_(  <_> )  |  
 |__|    \___  >____  > |__| |   __/|__|____/\____/|__|  
             \/     \/       |__|                        

`,
}

func GetRandomAsciiArt() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(asciiArtPresets))
	return asciiArtPresets[randomIndex]
}
