# nonamedreturns

I hate named returns in golang because they are error prone. That's why I wrote this linter. That's all

Tutorial on how to write your own linter:
https://disaev.me/p/writing-useful-go-analysis-linter/

Named errors used in defers are not reported. If you also want to report them set `report-error-in-defer` to true.