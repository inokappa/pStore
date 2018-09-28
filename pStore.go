package main

import (
    "os"
    "fmt"
    "flag"
    "strings"
    "bytes"
    "encoding/csv"
    "encoding/json"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
    "github.com/aws/aws-sdk-go/service/sts"
    "github.com/aws/aws-sdk-go/service/ssm"

    "github.com/olekukonko/tablewriter"
)

const (
    AppVersion = "0.0.1"
)

var (
    argProfile = flag.String("profile", "", "Profile 名を指定.")
    argRole = flag.String("role", "", "Role ARN を指定.")
    argRegion = flag.String("region", "ap-northeast-1", "Region 名を指定.")
    argEndpoint = flag.String("endpoint", "", "AWS API のエンドポイントを指定.")
    argVersion = flag.Bool("version", false, "バージョンを出力.")
    argCsv = flag.Bool("csv", false, "CSV 形式で出力する")
    argJson = flag.Bool("json", false, "JSON 形式で出力する")
    argPut = flag.Bool("put", false, "パラメータを追加する")
    argName = flag.String("name", "", "パラメータの名前を指定する")
    argValue = flag.String("value", "", "パラメータ名を値を指定する")
    argOverwrite = flag.Bool("overwrite", false, "パラメータを上書きする")
    argSecure = flag.Bool("secure", false, "SecureString でパラメータを追加する")
    argList = flag.Bool("list", false, "StringList でパラメータを追加する")
    argDel = flag.Bool("del", false, "パラメータを削除する")
)

type Parameters struct {
    Parameters      []Parameter    `json:"parameters"`
}

type Parameter struct {
    Name             string `json:"name"`
    Value            string `json:"value"`
    Type             string `json:"type"`
    LastModifiedDate string `json:"last_modified_date"`
}

func outputTbl(data [][]string) {
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Name", "Value", "Type", "LastModifiedDate"})
    for _, value := range data {
        table.Append(value)
    }
    table.Render()
}

func outputCsv(data [][]string) {
    buf := new(bytes.Buffer)
    w := csv.NewWriter(buf)
    for _, record := range data {
        if err := w.Write(record); err != nil {
            fmt.Println("Write error: ", err)
            return
        }
        w.Flush()
    }
    fmt.Println(buf.String())
}

func outputJson(data [][]string) {
    var rs []Parameter
    for _, record := range data {
        r := Parameter{Name:record[0], Value:record[1], Type:record[2],
                       LastModifiedDate:record[3]}
        rs = append(rs, r)
    }
    rj := Parameters{
        Parameters: rs,
    }
    b, err := json.Marshal(rj)
    if err != nil {
        fmt.Println("JSON Marshal error:", err)
        return
    }
    os.Stdout.Write(b)
}

func awsSsmClient(profile string, region string, role string) *ssm.SSM {
    var config aws.Config
    if profile != "" {
        creds := credentials.NewSharedCredentials("", profile)
        config = aws.Config{Region: aws.String(region),
                            Credentials: creds,
                            Endpoint: aws.String(*argEndpoint)}
    } else if role != "" {
        sess := session.Must(session.NewSession())
        assumeRoler := sts.New(sess)
        creds := stscreds.NewCredentialsWithClient(assumeRoler, role)
        config = aws.Config{Region: aws.String(region),
            Credentials: creds,
            Endpoint: aws.String(*argEndpoint)}
    } else {
        config = aws.Config{Region: aws.String(region),
                            Endpoint: aws.String(*argEndpoint)}
    }

    sess := session.New(&config)
    ssmClient := ssm.New(sess)
    return ssmClient
}

func putParameter(ssmClient *ssm.SSM, pName string, pType string, pValue string) {
    params := &ssm.PutParameterInput {
        Name: aws.String(pName),
        Value: aws.String(pValue),
        Description: aws.String(pName),
        Type: aws.String(pType),
    }
    if *argOverwrite {
        params.SetOverwrite(true)
    }

    _, err := ssmClient.PutParameter(params)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
}

func delParameter(ssmClient *ssm.SSM, pName string) {
    params := &ssm.DeleteParameterInput {
        Name: aws.String(pName),
    }
    _, err := ssmClient.DeleteParameter(params)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
}

func listParameters(ssmClient *ssm.SSM) {
    params := &ssm.DescribeParametersInput {}

    allParameters := [][]string{}
    for {
        res, err := ssmClient.DescribeParameters(params)
        if err != nil {
            fmt.Println(err.Error())
            os.Exit(1)
        }
        for _, r := range res.Parameters {
            // 別関数に切り出す
            const layout = "2006-01-02 15:04:05"
            jst := time.FixedZone("Asia/Tokyo", 9*60*60)
            d := *r.LastModifiedDate

            // 別関数に切り出す
            params := &ssm.GetParameterInput {
                Name: aws.String(*r.Name),
                WithDecryption: aws.Bool(true),
            }
            res, err := ssmClient.GetParameter(params)
            if err != nil {
                fmt.Println(err.Error())
                os.Exit(1)
            }

            var pValue string
            if *res.Parameter.Type == "SecureString" {
                pValue = "******************"
            } else {
                pValue = *res.Parameter.Value
            }

            Parameter := []string{
                *r.Name,
                pValue,
                *r.Type,
                d.In(jst).Format(layout),
            }
            allParameters = append(allParameters, Parameter)
        }
        if res.NextToken == nil {
            break
        }
        params.SetNextToken(*res.NextToken)
        continue
    }

    if *argCsv == true {
        outputCsv(allParameters)
    } else if *argJson == true {
        outputJson(allParameters)
    } else {
        outputTbl(allParameters)
    }
}

func main() {
    flag.Parse()

    if *argVersion {
      fmt.Println(AppVersion)
      os.Exit(0)
    }

    ssmClient := awsSsmClient(*argProfile, *argRegion, *argRole)

    if *argPut {
        if *argName == "" {
            fmt.Println("パラメータの名前を指定して下さい.")
            os.Exit(1)
        }
        // Type を選択する (デフォルトは String とする)
        var pType string
        if *argSecure {
            pType = "SecureString"
        } else if *argList || strings.Contains(*argName, "/") {
            pType = "StringList"
        } else {
            pType = "String"
        }
        putParameter(ssmClient, *argName, pType, *argValue)
    } else if *argDel {
        if *argName == "" {
            fmt.Println("パラメータの名前を指定して下さい.")
            os.Exit(1)
        }
        delParameter(ssmClient, *argName)
    } else {
        listParameters(ssmClient)
    }
}
