package configfile

var LogConfig = `
<seelog type="asynctimer" asyncinterval="5000000" minlevel="debug" maxlevel="error">
  <outputs formatid="normal">
  <!-- <console/> -->

        <filter levels="debug" formatid="normal">
            <rollingfile  type="size" filename="./log/log.log" maxsize="5242880" maxrolls="5" />
        </filter>

        <filter levels="error" formatid="error">
            <file path="./log/error.log"/>
        </filter>

        <filter levels="info" formatid="info">
            <rollingfile  type="size" filename="./log/info.log" maxsize="10485760" maxrolls="5" />
        </filter>

    </outputs>
    <formats>
	<format id="test" format='{"time": %Ns, "level": "%LEV", "relfile": "%RelFile", "line": "%Line", "msg": "%Msg"}%n'/>
        <format id="usetags" format='{"time": %Time, "msg": "%Msg"}%n'/>
        <format id="normal" format='time: %Ns, level: %LEV, relfile: %RelFile, line: %Line, msg: %Msg%n'/>
        <format id="info" format='{"time": %Ns, "level": "%LEV", "relfile": "%RelFile", "line": "%Line", "msg": %Msg}%n'/>
        <format id="error" format='{"time": %Ns, "level": "%LEV", "relfile": "%RelFile", "line": "%Line", "msg": %Msg}%n'/>
    </formats>
</seelog>
`
