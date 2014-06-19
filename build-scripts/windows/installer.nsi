!include nsDialogs.nsh
!include LogicLib.nsh

!define APPNAME "lsleases"


# installer location
Outfile "$%BUILD_OUTPUT%\$%BUILD_ARCH%\lsleases_installer_$%VERSION%_$%BUILD_ARCH%.exe"

InstallDir $PROGRAMFILES\${APPNAME}

LicenseData LICENSE

# installer title bar in installer / uninstaller
Name ${APPNAME}

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


  ${NSD_CreateCheckBox} 0 -80 100% 12u "autostart lsleases server on boot"
  pop $autostartCheckbox
  GetFunctionAddress $0 OnClick
  nsDialogs::OnClick $autostartCheckbox $0


  ${NSD_CreateLabel} 0 -50 100% 12u "to start an lsleases server instance go to: 'Start/Programms/lsleases/start server'"
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

  SetOutPath $INSTDIR

  File "lsleases.exe"
  File "manual.html"
  File "manual.txt"
  File "nssm.exe"
  File "list-leases.bat"
  File "clear-leases.bat"
  File "start-server.bat"
  File "stop-server.bat"

  # start menu
  createDirectory "$SMPROGRAMS\${APPNAME}"
  createShortCut  "$SMPROGRAMS\${APPNAME}\list leases.lnk" "$INSTDIR\list-leases.bat"
  createShortCut  "$SMPROGRAMS\${APPNAME}\clear leases.lnk" "$INSTDIR\clear-leases.bat"
  createShortCut  "$SMPROGRAMS\${APPNAME}\uninstall.lnk" "$INSTDIR\uninstall.exe"

  ${If} $autostartCheckboxState == ${BST_CHECKED}
    # register service per nssm wrapper
    ExecWait '"$INSTDIR\nssm.exe" install ${APPNAME} "$INSTDIR\start-server.bat"'
    Sleep 1000
    Exec '"$INSTDIR\nssm.exe" start ${APPNAME}'
  ${Else}  
    # install start / stop server link in start menu
    createShortCut  "$SMPROGRAMS\${APPNAME}\start server.lnk" "$INSTDIR\start-server.bat"
    createShortCut  "$SMPROGRAMS\${APPNAME}\stop server.lnk" "$INSTDIR\stop-server.bat"
  ${EndIf}

  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd


# uninstaller

function un.onInit

  MessageBox MB_OKCANCEL "uninstall ${APPNAME}?" IDOK next
    abort
  next:
functionEnd

Section "uninstall"
  # stop server instance
  ExecWait "$INSTDIR\stop-server.bat"
  Sleep 500

  # remove service
  ExecWait "$INSTDIR\nssm.exe remove ${APPNAME} confirm"
  Sleep 1000


  # start menu
  delete "$SMPROGRAMS\${APPNAME}\list leases.lnk"
  delete "$SMPROGRAMS\${APPNAME}\clear leases.lnk"
  delete "$SMPROGRAMS\${APPNAME}\uninstall.lnk"
  delete "$SMPROGRAMS\${APPNAME}\start server.lnk"
  delete "$SMPROGRAMS\${APPNAME}\stop server.lnk"

  rmDir  "$SMPROGRAMS\${APPNAME}"

  # programm files
  rmDir /r "$INSTDIR\*.*"
  rmDir "$INSTDIR"

SectionEnd
