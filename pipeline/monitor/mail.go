package monitor

import (
	"github.com/go-gomail/gomail"
	"crypto/tls"
	"log"
)

func Sendmail(err error) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "maixiang@jd.com", "端口扫描2.0") // 发件人
	m.SetHeader("To", // 收件人
		m.FormatAddress("wangshuo11@jd.com", "wangshuo"),
	)

	m.SetHeader("Subject", "安全报警报告-报警报告-goSkylar") // 主题

	body := `<!DOCTYPE html>
<html>

	<head>
		<meta charset="UTF-8">
		<title></title>
	</head>

	<body>
		<div id="qm_con_body">
			<div id="mailContentContainer" class="qmbox qm_con_body_content qqmail_webmail_only" style="">

				<style type="text/css">
					.qmbox #outlook a {
						padding: 0;
					}

					.qmbox .ReadMsgBody {
						width: 100%;
					}

					.qmbox .ExternalClass {
						width: 100%;
					}

					.qmbox body {
						-webkit-text-size-adjust: 100%;
						-ms-text-size-adjust: 100%;
						-webkit-font-smoothing: antialiased;
					}

					.qmbox .yshortcuts,
					.qmbox .yshortcuts a,
					.qmbox .yshortcuts a:link,
					.qmbox .yshortcuts a:visited,
					.qmbox .yshortcuts a:hover,
					.qmbox .yshortcuts a span {
						text-decoration: none !important;
						border-bottom: none !important;
						background: none !important;
					}
					/*link style*/

					.qmbox a {
						color: #43a0dd;
						text-decoration: none;
						outline: none;
					}

					.qmbox .top-bar a {
						color: #abb4bd;
					}

					.qmbox .menu a {
						color: #77818c;
					}

					.qmbox .home a {
						color: #43a0dd;
					}
					/*.button a { color:#ffffff; }*/

					.qmbox a:hover {
						text-decoration: underline !important;
					}

					.qmbox .top-bar a:hover {
						text-decoration: none !important;
					}

					.qmbox .button a:hover,
					.qmbox .button2 a:hover {
						text-decoration: none !important;
					}

					@media only screen and (max-width: 640px) {
						.qmbox table[class~=wrap],
						.qmbox table[class~=divider] {
							width: 100% !important;
						}
					}

					@media only screen and (max-width: 800px) {
						.qmbox table[class~=wrap],
						.qmbox table[class~=divider] {
							width: 440px !important;
						}
						.qmbox table[class~=wrap][class~=color],
						.qmbox table[class~=wrap][class~=header-img] {
							width: 450px !important;
							-webkit-box-shadow: 0 0 10px rgba(0, 0, 0, 0.6);
							-moz-box-shadow: 0 0 10px rgba(0, 0, 0, 0.6);
							box-shadow: 0 0 10px rgba(0, 0, 0, 0.6);
						}
						.qmbox table[class~=row] {
							width: 400px !important;
						}
						.qmbox table[class~=header] .in {
							padding-top: 30px !important;
						}
						.qmbox table[class~=header-img] .in {
							padding: 5px !important;
						}
						.qmbox td[class~=header-img-td] {
							padding: 100px 35px !important;
						}
						.qmbox table[class~=avator-box] {
							width: 100% !important;
							float: none !important;
						}
						.qmbox table[class~=col2] {
							width: 100% !important;
						}
						.qmbox table[class~=col3] {
							width: 100% !important;
						}
						.qmbox table[class~=col-1-3] {
							width: 100% !important;
						}
						.qmbox table[class~=col-2-3] {
							width: 100% !important;
						}
						.qmbox table[class=footer-left] {
							width: 100% !important;
						}
						.qmbox table[class=footer-right] {
							width: 100% !important;
						}
						.qmbox td[class~=general-td] {
							padding: 10px 10px 0 10px !important;
						}
						.qmbox td[class~=general-img-td] {
							padding: 10px !important;
						}
						.qmbox table[class=footer-left] td {
							text-align: center !important;
						}
						.qmbox table[class=footer-right] td {
							text-align: center !important;
						}
						.qmbox table[class~=bottom] .row {
							text-align: center !important;
						}
						.qmbox table[class~=bottom] .col3 {
							width: 80% !important;
							margin: 0 auto !important;
							float: none !important;
						}
						.qmbox table[class~=bottom] .col3 .content {
							text-align: center !important;
						}
						.qmbox table[class~=footer] .in {
							padding: 15px 0 5px !important;
						}
						.qmbox table[class~=social-icons-t] {
							float: none !important;
							margin: 0 auto !important;
						}
						.qmbox table[class~=mid] {
							float: none !important;
							margin-bottom: 25px !important;
						}
						.qmbox table[class~=logo],
						.qmbox table[class~=menu],
						.qmbox table[class~=menu2] {
							width: 90% !important;
							float: none !important;
							margin: 0 auto 10px !important;
						}
						.qmbox table[class~=logo] {
							margin: 0 auto 30px !important;
						}
						.qmbox table[class~=menu] .info {
							display: none !important;
						}
						.qmbox table[class~=menu] td {
							text-align: center !important;
							font-size: 13px !important;
						}
						.qmbox img {
							height: auto !important;
						}
						.qmbox img[class~=img] {
							width: 100% !important;
							height: auto !important;
							max-width: 100% !important;
							display: block !important;
						}
						.qmbox table[class~=menu2] .in {
							text-align: center !important;
						}
					}

					@media only screen and (max-width: 449px) {
						.qmbox table[class~=wrap],
						.qmbox table[class~=divider],
						.qmbox table[class~=wrap][class~=color],
						.qmbox table[class~=wrap][class~=header-img] {
							width: 100% !important;
						}
						.qmbox table[class~=row] {
							width: 100% !important;
						}
						.qmbox table[class~=logo] img {
							max-width: 100% !important;
						}
						.qmbox table[class~=logo],
						.qmbox table[class~=menu],
						.qmbox table[class~=menu2] {
							width: 97% !important;
						}
						.qmbox table[class~=title-box] {
							width: 100% !important;
						}
						.qmbox table[class~=title-box] td {
							text-align: center !important;
						}
						.qmbox table[class~=avator-box],
						.qmbox table[class~=quote-box] {
							width: 100% !important;
						}
						.qmbox td[class~=header-img-td] h2 {
							font-size: 18px !important;
							line-height: 24px !important;
						}
					}

					@media only screen and (max-width: 339px) {
						.qmbox table[class~=logo] img {
							max-width: 260px !important;
						}
						.qmbox table[class~=full] {
							width: 100% !important;
						}
						.qmbox table[class~=full] .in {
							padding: 0 0 20px 0 !important;
						}
						.qmbox table[class~=full] img {
							width: 100% !important;
						}
					}

					.qmbox .STYLE1 {
						font-size: 20px;
					}

					.qmbox .STYLE5 {
						font-size: 14px;
					}
				</style>

				<table class="BGtable" style=" border-collapse: collapse; margin: 0px; padding: 0px; background-color: #f8f8fa; width: 100%; height: 100%; background-image: url(); " background="" border="0" cellpadding="0" cellspacing="0" width="100%">
					<tbody>
						<tr>
							<td class="BGtable-inner" valign="top">
								<table class="wrap top-bar" style="border-collapse: collapse; background-color: #34c4fd; width: 800px; margin: 0px auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
									<tbody>
										<tr>
											<td class="in" style="padding: 0;">
												<table class="row" style="border-collapse: collapse;width: 800px;margin: 0 auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="600">
													<tbody>
														<tr>
															<td class="in" style="padding: 0 10px;" valign="top">
																<table style="border-collapse: collapse;border: none;mso-table-lspace: 0pt;mso-table-rspace: 0pt;" align="left" border="0" cellpadding="0" cellspacing="0">
																	<tbody>
																		<tr>
																			<td class="content" style="font-family: Tahoma, Arial;font-size: 14px;line-height: 13px;font-weight: 400;color: #FFFFFF;-webkit-font-smoothing: antialiased;padding: 8px 0 9px;">
																				<singleline label="top-bar">
																					<webversion>
																						信息安全部－脉像平台
																					</webversion>
																				</singleline>
																			</td>
																		</tr>
																	</tbody>
																</table>
															</td>
														</tr>
													</tbody>
												</table>
											</td>
										</tr>
									</tbody>
								</table>
								<layout label="header-1s1">
									<table class="wrap header" style="border-collapse: collapse; background-color: #ffffff; width: 800px; margin: 0px auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
										<tbody>
											<tr>
												<td class="in" style="padding: 25px 0 20px;">
													<table class="row" style="border-collapse: collapse;width: 740px;margin: 0 auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="600">
														<tbody>
															<tr>
																<td class="general-img-td" style="padding: 10px;" valign="top">

																	<table class="logo mid" style="border-collapse: collapse;border: none;mso-table-lspace: 0pt;mso-table-rspace: 0pt;" align="left" border="0" cellpadding="0" cellspacing="0">
																		<tbody>
																			<tr>
																				<td width="120" align="center" valign="top"> <img src="http://sea.jd.com/static/picture/pulseLogo.png" style="border:0 none;" height="70" width="70"> </td>
																				<td width="503" height="44" align="left" valign="center" style="font-family: 'Open Sans',Arial,Tohama; color: #4e5762; font-weight: 300;font-size:30px;">脉像 • 谛听</td>
																			</tr>
																		</tbody>
																	</table>
																</td>
															</tr>
														</tbody>
													</table>
												</td>
											</tr>
										</tbody>
									</table>
								</layout>
								<layout label="col2-right-news2">
									<table class="wrap" style="border-collapse: collapse; background-color: #ffffff; width: 800px; margin: 0px auto; background-image: none;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
										<tbody>
											<tr>
												<td class="in">
													<table class="col2" style="border-collapse: collapse;width: 95%;border: none;mso-table-lspace: 0pt;mso-table-rspace: 0pt;" align="right" border="0" cellpadding="0" cellspacing="0" width="295">
														<tbody>
															<tr>
																<td class="general-img-td content" style="font-family: Tahoma,Arial; font-size: 13px; line-height: 21px; font-weight: 400; color: #929ba5; padding: 10px;">
																	<h3 class="h3" style="font-family: 'Open Sans',Arial,Tohama; color: #4e5762; font-weight: 300; font-size: 18px; line-height: 24px; margin: 0px 0px 10px;"> <span class="STYLE1">
                 <singleline label="h3-title2">
                  <strong></strong>
                 </singleline> </span> </h3>
																	<table style="border-collapse: collapse;" border="0" cellpadding="0" cellspacing="0" width="100%">
																		<tbody>
																			<tr>
																				<td class="content border-bottom" style="font-family: Tahoma,Arial; font-size: 15px; line-height: 21px; font-weight: 400; color: #929ba5; border-bottom: 1px dotted #e5e5e5;padding: 8px 0px 0px;">
																					<table class="img-rounded" style="border-collapse: collapse;border: none;mso-table-lspace: 0pt;mso-table-rspace: 0pt;" align="left" border="0" cellpadding="0" cellspacing="0">
																						<tbody></tbody>
																					</table>`

	body = body + `<table width="80%" border="0" cellspacing="0" cellpadding="0">
																						<tbody>
																							<tr>
																								<td>
																									<table width="710" border="0" cellpadding="0" cellspacing="0" style="border-top:1px solid #f0f0f0; padding-top:8px; padding-bottom:8px;">
																										<tbody>
																											<tr>
																												<td height="36" colspan="2">
																													<div align="left">`
	body = body + err.Error()

	body = body + `</td>
																			</tr>
																		</tbody>
																	</table>
																</td>
															</tr>
														</tbody>
													</table>
												</td>
											</tr>
										</tbody>
									</table>
								</layout>

								<layout label="space-25px2">
									<table class="divider" style="border-collapse: collapse;width: 800px;margin: 0 auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
										<tbody>
											<tr>
												<td class="wrap row1" style="background-color: #ffffff; width: 800px; margin: 0px auto; background-image: none;" height="25"></td>
											</tr>
										</tbody>
									</table>
								</layout>
								<layout label="space-25px1">
									<table class="divider" style="border-collapse: collapse;width: 800px;margin: 0 auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
										<tbody>
											<tr>
												<td class="wrap row1" style="background-color: #ffffff; width: 800px; margin: 0px auto; background-image: none;" height="25"></td>
											</tr>
										</tbody>
									</table>
								</layout>
								<table class="wrap footer" style=" border-collapse: collapse; background-color: #34c4fd; background-image: url(); width: 800px; margin: 0px auto; " align="center" background="" border="0" cellpadding="0" cellspacing="0" width="800">
									<tbody>
										<tr>
											<td class="in" style="padding: 6px 0;">
												<table class="row" style="border-collapse: collapse;width: 800px;margin: 0 auto;" align="center" border="0" cellpadding="0" cellspacing="0" width="800">
													<tbody>
														<tr>
															<td class="in" style="padding: 0 10px;" align="center" height="45">
																<table class="footer-left" style="border-collapse: collapse;border: none;mso-table-lspace: 0pt;mso-table-rspace: 0pt;" align="right" border="0" cellpadding="0" cellspacing="0">
																	<tbody>
																		<tr>
																			<td style="font-family: Tahoma, Arial; font-size: 12px; line-height: 13px; font-weight: 400; color: #FFFFFF; -webkit-font-smoothing: antialiased; text-align: left; padding: 10px 0px 0px 0px;">信息安全部 － 脉象平台</td>
																		</tr>
																		<tr>
																			<td class="content" style="font-family: Tahoma, Arial; font-size: 12px; line-height: 13px; font-weight: 400; color: #FFFFFF; -webkit-font-smoothing: antialiased; text-align: left; padding: 8px 0px 9px;">
																				<singleline label="copyright">
																					Copyright © 2017 -
																					<currentyear>
																						<a href="http://pulse.jd.com" target="_blank" style="color: #FFFFFF;">pulse.jd.com</a>
																					</currentyear>
																				</singleline>
																			</td>
																		</tr>
																	</tbody>
																</table>
															</td>
														</tr>
													</tbody>
												</table>
											</td>
										</tr>
									</tbody>
								</table>
							</td>
						</tr>
					</tbody>
				</table>

				<style type="text/css">
					.qmbox style,
					.qmbox script,
					.qmbox head,
					.qmbox link,
					.qmbox meta {
						display: none !important;
					}
				</style>
			</div>
		</div>
	</body>

</html>`

	m.SetBody("text/html", body) // 正文

	d := gomail.NewPlainDialer("mx.jd.local", 25, "360buyAD.local/maixiang", "RWqwzyxsuyu*384") // 发送邮件服务器、端口、发件人账号、发件人密码

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	} else {
		log.Println("Emergency 监控 Mail 已发送")
	}
}
