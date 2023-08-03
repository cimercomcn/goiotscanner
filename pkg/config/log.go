package config

import (
    "github.com/fatih/color"
    "github.com/neumannlyu/golog"
)

// 日志对象集
type LogSet struct {
    // Fg:Green
    // Tag:[ OK ]
    OK golog.SimpleLog
    // Tag: [   OK    ] FgHiBlue Underline
    OK2 golog.SimpleLog
    // Fg:Green
    // Tag:[PASS]
    Pass golog.SimpleLog
    // Fg:Green
    // Tag:[IGNORED]
    // Underline
    Ignored golog.SimpleLog
    // Fg:Blue
    // Tag:[SKIP]
    // Underline
    Skip golog.SimpleLog
    // Tag: [UNKNOWN] Fg
    Unknwon   golog.SimpleLog
    CommonLog golog.CommonLog
}

func (logset *LogSet) applyDefault() {
    golog.UnifiedLogData.Fgcolor = color.FgGreen
    golog.UnifiedLogData.FormatString = "[" +
        golog.UnifiedLogData.FormatString + "]"

    // OK
    logset.OK = golog.NewSimpleLog()
    logset.OK.Tag.Fgcolor = color.FgGreen
    logset.OK.Tag.Bgcolor = 0
    logset.OK.Tag.FormatString = " OK "
    logset.OK.FormatString = golog.UnifiedLogFormatString
    // OK2
    logset.OK2 = golog.NewSimpleLog()
    logset.OK2.Tag.Fgcolor = color.FgHiBlue
    logset.OK2.Tag.Bgcolor = 0
    logset.OK2.Tag.Font = color.Underline
    logset.OK2.Tag.FormatString = "      OK "
    logset.OK2.FormatString = "                     "
    // Pass
    logset.Pass = golog.NewSimpleLog()
    logset.Pass.Tag.Fgcolor = color.FgGreen
    logset.Pass.Tag.Bgcolor = 0
    logset.Pass.Tag.FormatString = "PASS"
    logset.Pass.FormatString = golog.UnifiedLogFormatString
    // Ignored
    logset.Ignored = golog.NewSimpleLog()
    logset.Ignored.Tag.Fgcolor = color.FgGreen
    logset.Ignored.Tag.Bgcolor = 0
    logset.Ignored.Tag.FormatString = "IGNORED"
    logset.Ignored.Tag.Font = color.Underline
    logset.Ignored.FormatString = golog.UnifiedLogFormatString
    // Skip
    logset.Skip = golog.NewSimpleLog()
    logset.Skip.Tag.Fgcolor = color.FgBlue
    logset.Skip.Tag.Bgcolor = 0
    logset.Skip.Tag.FormatString = "SKIP"
    logset.Skip.Tag.Font = color.Underline
    logset.Skip.Msg.Bgcolor = color.BgWhite
    logset.Skip.Msg.Fgcolor = color.FgYellow
    logset.Skip.FormatString = golog.UnifiedLogFormatString
    // Unknwon
    logset.Unknwon = golog.NewSimpleLog()
    logset.Unknwon.Tag.Fgcolor = color.FgYellow
    logset.Unknwon.Tag.Bgcolor = 0
    logset.Unknwon.Tag.FormatString = "UNKNOWN"
    logset.Unknwon.Tag.Font = color.Underline
    logset.Unknwon.FormatString = golog.UnifiedLogFormatString
    // CommonLog
    logset.CommonLog = golog.NewCommonLog()
    logset.CommonLog.Format = golog.UnifiedLogFormatString
}
