/*===============================================================
*   Copyright (C) 2020 All rights reserved.
*
*   FileName：plugin.go
*   Author：WuGuoFu
*   Date： 2020-07-30
*   Description：
*
================================================================*/
package confutil

//register function
//anyFileMap key is the plugin name which provide by flag args "-c=any"
//anyFileMap value is the plugin load function,such as the loadAny function in goany.go
func loadPlugin() {
	anyFileMap["any"] = loadAny
}
