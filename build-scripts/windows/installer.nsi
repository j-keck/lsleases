!include nsDialogs.nsh
!include LogicLib.nsh

!define APPNAME "lsleases"

RequestExecutionLevel admin

# installer location
Outfile "$%BUILD_OUTPUT%\lsleases_$%VERSION%_win_installer_$%BUILD_ARCH%.exe"

InstallDir $PROGRAMFILES\${APPNAME}

LicenseData LICENSE

# installer title bar in installer / uninstaller
Name "${APPNAME} Version: $%VERSION%"

var customDialogPage
var welcomeLabel
var descriptionLabel
var howToStartServerInstanceLabel
var autostartCheckboxDisabledDescriptionLabel
var autostartCheckbox
var autostartCheckboxState
var currentUserAccountType

page license
page custom CustomDialogPage CustomDialogPageLeave
page directory 
page instfiles

Function CustomDialogPage
  nsDialogs::Create 1018
  Pop $customDialogPage
  ${If} customDialogPage == error
    Abort
  ${EndIf}


  ${NSD_CreateLabel} 0 0 100% 15u "lsleases - dhcp leases sniffer"
  pop $welcomeLabel
  CreateFont $0 "Arial" 12
  SendMessage $welcomeLabel ${WM_SETFONT} $0 1


  ${NSD_CreateLabel} 0 -110 100% 12u "! enable autostart lsleases server not possible - needs admin rights !"
  pop $autostartCheckboxDisabledDescriptionLabel
  CreateFont $0 "Arial" 10
  SendMessage $autostartCheckboxDisabledDescriptionLabel ${WM_SETFONT} $0 1
  ShowWindow $autostartCheckboxDisabledDescriptionLabel ${SW_HIDE} # hidden by default


  ${NSD_CreateCheckBox} 0 -90 100% 12u "autostart lsleases server on boot"
  pop $autostartCheckbox
  GetFunctionAddress $0 OnClick
  nsDialogs::OnClick $autostartCheckbox $0


  ${NSD_CreateLabel} 0 -60 100% 15u "to start an lsleases server instance go to: 'Start/Programms/lsleases/start server'"
  pop $howToStartServerInstanceLabel
  ShowWindow $howToStartServerInstanceLabel ${SW_HIDE} # hidden by default



  ${NSD_CreateLabel} 0 -20 100% 60u "to list leases go to: 'Start/Programms/lsleases/list leases'"
  pop $descriptionLabel



  # check current user
  UserInfo::GetOriginalAccountType
  pop $currentUserAccountType

  ${If} $currentUserAccountType == "Admin"
    ${NSD_Check} $autostartCheckbox
  ${Else}
    EnableWindow $autostartCheckbox 0

    ShowWindow $autostartCheckboxDisabledDescriptionLabel ${SW_SHOW}

    ShowWindow $howToStartServerInstanceLabel ${SW_SHOW}
  ${EndIf}  



  nsDialogs::Show
FunctionEnd

Function CustomDialogPageLeave
  pop $0

  ${NSD_GetState} $autostartCheckbox $autostartCheckboxState
FunctionEnd


Function OnClick
  pop $0
  
  ${NSD_GetState} $0 $1
  ${If} $1 == 1
    ShowWindow $howToStartServerInstanceLabel ${SW_HIDE}
  ${Else}
    ShowWindow $howToStartServerInstanceLabel ${SW_SHOW}
  ${EndIf}  
FunctionEnd  
    


Section "install"
  SetShellVarContext all

  SetOutPath $INSTDIR

  File "lsleases.exe"
  File "manual.html"
  File "nssm.exe"
  File "list-leases.bat"
  File "watch-leases.bat"
  File "clear-leases.bat"
  
  # start menu
  createDirectory "$SMPROGRAMS\${APPNAME}"
  createShortCut  "$SMPROGRAMS\${APPNAME}\list leases.lnk" "$INSTDIR\list-leases.bat"
  createShortCut  "$SMPROGRAMS\${APPNAME}\watch for new leases.lnk" "$INSTDIR\watch-leases.bat"
  createShortCut  "$SMPROGRAMS\${APPNAME}\clear leases.lnk" "$INSTDIR\clear-leases.bat"
  createShortCut  "$SMPROGRAMS\${APPNAME}\uninstall.lnk" "$INSTDIR\uninstall.exe"
  createShortCut  "$SMPROGRAMS\${APPNAME}\manual.lnk" "$INSTDIR\manual.html"

  ${If} $autostartCheckboxState == ${BST_CHECKED}
    # register service per nssm wrapper
    ExecWait '"$INSTDIR\nssm.exe" install ${APPNAME} "$INSTDIR\lsleases.exe" -s'
    Sleep 1000
    Exec '"$INSTDIR\nssm.exe" start ${APPNAME}'

    # service controller scripts
    File "start-service.bat"
    File "stop-service.bat"
    File "restart-service.bat"

  ${Else}
    # server start / stop scripts
    File "start-server.bat"
    File "stop-server.bat"
  
    # install start / stop server link in start menu
    createShortCut  "$SMPROGRAMS\${APPNAME}\start server.lnk" "$INSTDIR\start-server.bat"
    createShortCut  "$SMPROGRAMS\${APPNAME}\stop server.lnk" "$INSTDIR\stop-server.bat"
  ${EndIf}

  # write installed flag in registry
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "curVer" $%VERSION%
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "instdir" $INSTDIR

  # add firwall rule
  Exec 'netsh advfirewall firewall add rule name=lsleases dir=in action=allow program="$INSTDIR\lsleases.exe" enable=yes'


  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

function .onInit 
  # check installed flag in registry
  ReadRegStr $R0 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "curVer"
  ${If} $R0 != "" 
    MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION \
    "${APPNAME} is already in version $R0 installed. $\n$\nClick 'OK' to remove the previous version or 'Cancel' to cancel this upgrade." \
    IDOK uninst
    Abort

    uninst:
      ReadRegStr $R1 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "instdir"
      ClearErrors
      ExecWait '"$R1\uninstall.exe" /S'
  ${EndIf}
functionEnd

# uninstaller

function un.onInit

  IfSilent next
  MessageBox MB_OKCANCEL "uninstall ${APPNAME}?" IDOK next
    abort
  next:
functionEnd

Section "uninstall"
  SetShellVarContext all

  # stop server instance
  ExecWait "$INSTDIR\stop-server.bat"
  ExecWait "$INSTDIR\stop-service.bat"
  Sleep 500

  # remove service
  ExecWait "$INSTDIR\nssm.exe remove ${APPNAME} confirm"
  Sleep 1000


  # start menu
  delete "$SMPROGRAMS\${APPNAME}\list leases.lnk"
  delete "$SMPROGRAMS\${APPNAME}\watch for new leases.lnk"
  delete "$SMPROGRAMS\${APPNAME}\clear leases.lnk"
  delete "$SMPROGRAMS\${APPNAME}\uninstall.lnk"
  delete "$SMPROGRAMS\${APPNAME}\start server.lnk"
  delete "$SMPROGRAMS\${APPNAME}\stop server.lnk"
  delete "$SMPROGRAMS\${APPNAME}\manual.lnk"

  rmDir  "$SMPROGRAMS\${APPNAME}"

  # remove installed flag in registry
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"

  # remove firewall rule
  Exec 'netsh advfirewall firewall delete rule name=lsleases program="$INSTDIR\lsleases.exe"'

  # programm files
  rmDir /r "$INSTDIR\*.*"
  rmDir "$INSTDIR"

SectionEnd
