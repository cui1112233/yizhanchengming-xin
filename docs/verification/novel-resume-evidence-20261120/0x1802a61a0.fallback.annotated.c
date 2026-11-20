/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802a61a0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

code * FUN_1802a61a0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  int *piVar2;
  int *piVar3;
  longlong *plVar4;
  char cVar5;
  int iVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  longlong *plVar12;
  code *pcVar13;
  code *pcVar14;
  undefined4 uVar15;
  undefined8 uVar16;
  undefined1 auStack_b8 [40];
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  int local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  code *local_50;
  code *local_48;
  ulonglong local_40;

  plVar4 = DAT_1803c3a28;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_b8;
  plVar1 = (longlong *)*param_3;
  local_90 = DAT_180391e78;
  if (DAT_1803c3a28 == (longlong *)0x0) {
LAB_1802a6212:
    DAT_1803c3a28 = (longlong *)FUN_18036cd10(param_1,DAT_1803c39b0,DAT_1803ca4b0,8);
  }
  else if ((1 < *DAT_1803c3a28) || (DAT_1803c3a28[2] != 0)) {
    if ((DAT_1803c3a28 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c3a28 && (*DAT_1803c3a28 = *DAT_1803c3a28 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1802a6212;
  }
  plVar4 = DAT_1803c3a28;
  FUN_180001d00(param_1,DAT_1803c3a28);
  plVar7 = (longlong *)FUN_18037f0c0();
  local_88 = plVar7;
  if (plVar7 == (longlong *)0x0) {
    plVar12 = *(longlong **)(param_1 + 0x60);
    uVar15 = 0x7a;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    plVar8 = (longlong *)FUN_18037f470();
    if (plVar8 == (longlong *)0x0) {
      plVar7 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar9 = (longlong *)0x0;
      plVar8 = (longlong *)0x0;
LAB_1802a6a9c:
      plVar12 = plVar7;
      uVar15 = 0x7a;
    }
    else {
      *(undefined4 *)(plVar4 + 5) = 0x7a;
      plVar9 = (longlong *)FUN_180376210(param_1,plVar8);
      local_60 = plVar9;
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      plVar8 = (longlong *)0x0;
      if (plVar9 == (longlong *)0x0) {
        plVar7 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar9 = (longlong *)0x0;
        goto LAB_1802a6a9c;
      }
      local_68 = plVar9;
      plVar10 = (longlong *)FUN_18037f470();
      local_80 = plVar10;
      if (plVar10 == (longlong *)0x0) {
        plVar7 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802a6a9c;
      }
      local_70 = 1;
      local_78 = plVar10;
      plVar8 = (longlong *)FUN_18037f0c0();
      if (plVar8 != (longlong *)0x0) {
        plVar9 = (longlong *)FUN_18037f0c0();
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar9 == (longlong *)0x0) {
          uVar16 = 0x7b;
          goto LAB_1802a669d;
        }
/* novels.services.scheduler#52="paused" */
        iVar6 = FUN_1803836d0(plVar9,DAT_1803c3c30);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if (iVar6 == -1) {
          uVar16 = 0x7b;
          goto LAB_1802a669d;
        }
        if (iVar6 == 1) {
          plVar8 = (longlong *)FUN_18037f0c0();
          if (plVar8 == (longlong *)0x0) {
            uVar16 = 0x7c;
            goto LAB_1802a669d;
          }
          cVar5 = FUN_18037f3a0();
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (cVar5 == '\0') {
            uVar16 = 0x7c;
            goto LAB_1802a669d;
          }
          plVar8 = (longlong *)FUN_18037f0c0();
          if (plVar8 == (longlong *)0x0) {
            uVar16 = 0x7d;
            goto LAB_1802a669d;
          }
          plVar9 = (longlong *)FUN_18037f0c0();
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          uVar16 = 0x7d;
          if (plVar9 == (longlong *)0x0) goto LAB_1802a669d;
          *(undefined4 *)(plVar4 + 5) = 0x7d;
          plVar8 = (longlong *)FUN_180376210(param_1,plVar9);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (plVar8 == (longlong *)0x0) goto LAB_1802a669d;
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          uVar16 = 0x7e;
          *(undefined4 *)(plVar4 + 5) = 0x7e;
/* novels.services.scheduler#41="_save_state" */
          plVar8 = (longlong *)FUN_18037e560(param_1,plVar1,DAT_1803c3bd8);
          if (plVar8 == (longlong *)0x0) goto LAB_1802a669d;
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          lVar11 = FUN_1802a0f10(param_1);
          if (lVar11 != 0) {
            plVar8 = (longlong *)FUN_18037f0c0();
            if (plVar8 == (longlong *)0x0) {
              uVar16 = 0x7f;
            }
            else {
              plVar9 = (longlong *)FUN_18037f0c0();
/* novels.services.scheduler#44="novel-scheduler" */
              pcVar13 = DAT_1803c3bf0;
              pcVar14 = _Py_TrueStruct_exref;
              if (plVar9 == (longlong *)0x0) {
                plVar9 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                }
                uVar16 = 0x7f;
                goto LAB_1802a66a9;
              }
              uVar16 = 0x7f;
              *(undefined4 *)(plVar4 + 5) = 0x7f;
              local_48 = pcVar13;
              local_50 = pcVar14;
              local_58 = plVar9;
/* novels.services.scheduler#45={"type": "tuple", "items": [{"tag": "a", "fileRange": [4367332, 4367340], "value": "target"}, {"tag": "a", "fileRange": [4367340, 4367348], "value": "daemon"}, {"tag": "a", "fileRange": [4367348, 4367354], "value": "name"}]} */
              plVar12 = (longlong *)FUN_18037ccb0(param_1,plVar8,&local_58,DAT_1803c3bf8);
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                (**(code **)(plVar9[1] + 0x30))(plVar9);
              }
              plVar7 = local_88;
              if (plVar12 != (longlong *)0x0) {
                cVar5 = FUN_18037f3a0();
                if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                  (**(code **)(plVar12[1] + 0x30))(plVar12);
                }
                plVar7 = local_88;
                if (cVar5 != '\0') {
                  plVar8 = (longlong *)FUN_18037f0c0();
                  uVar16 = 0x80;
                  plVar7 = local_88;
                  if (plVar8 != (longlong *)0x0) {
                    *(undefined4 *)(plVar4 + 5) = 0x80;
/* novels.services.scheduler#46="start" */
                    plVar9 = (longlong *)FUN_18037e560(param_1,plVar8,DAT_1803c3c00);
                    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                      (**(code **)(plVar8[1] + 0x30))(plVar8);
                    }
                    plVar7 = local_88;
                    if (plVar9 != (longlong *)0x0) {
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                        plVar7 = local_88;
                      }
                      goto LAB_1802a6a6d;
                    }
                  }
                }
              }
            }
            goto LAB_1802a669d;
          }
/* novels.services.scheduler#3="threading" */
          FUN_180373340(param_1,&local_90,DAT_1803c3aa8);
          uVar16 = 0x7f;
          plVar9 = local_90;
          goto LAB_1802a66a9;
        }
LAB_1802a6a6d:
        *(undefined4 *)(plVar4 + 5) = 0x7a;
/* novels.services.scheduler#47={"type": "tuple", "items": [{"tag": "n", "fileRange": [4367363, 4367364], "value": null}, {"tag": "n", "fileRange": [4367364, 4367365], "value": null}, {"tag": "n", "fileRange": [4367365, 4367366], "value": null}]} */
        plVar8 = (longlong *)FUN_1803793d0(param_1,plVar10,DAT_1803c3c08);
        if (plVar8 == (longlong *)0x0) {
          plVar7 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar9 = local_68;
          plVar8 = plVar10;
          goto LAB_1802a6a9c;
        }
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
LAB_1802a68ae:
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        if ((-1 < (int)*local_60) && (*local_60 = *local_60 + -1, *local_60 == 0)) {
          (**(code **)(local_68[1] + 0x30))();
        }
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(local_78[1] + 0x30))();
        }
        uVar15 = 0x81;
        *(undefined4 *)(plVar4 + 5) = 0x81;
        plVar7 = (longlong *)
/* novels.services.scheduler#48="_emit" ; novels.services.scheduler#54={"type": "tuple", "items": [{"tag": "u", "fileRange": [4367450, 4367469], "value": "scheduler:resumed"}]} */
                 FUN_18037e6c0(param_1,plVar1,DAT_1803c3c10,*(undefined8 *)(DAT_1803c3c40 + 0x18));
        if (plVar7 != (longlong *)0x0) {
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          FUN_180001d80(param_1);
          pcVar14 = _Py_NoneStruct_exref;
          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
            *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
          }
          if ((int)*plVar1 < 0) {
            return pcVar14;
          }
          *plVar1 = *plVar1 + -1;
          if (*plVar1 != 0) {
            return pcVar14;
          }
          (**(code **)(plVar1[1] + 0x30))(plVar1);
          return pcVar14;
        }
        plVar12 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1802a693e;
      }
      uVar16 = 0x7b;
LAB_1802a669d:
      plVar9 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1802a66a9:
      piVar3 = *(int **)(param_1 + 0xf8);
      local_90 = (longlong *)0x0;
      if ((piVar3 != (int *)0x0) && (*piVar3 + 1 != 0)) {
        *piVar3 = *piVar3 + 1;
      }
      piVar2 = (int *)plVar9[5];
      pcVar14 = (code *)FUN_18038e2a0(plVar4,uVar16);
      if (piVar2 == (int *)0x0) {
        plVar8 = (longlong *)plVar9[5];
        if (pcVar14 == _Py_NoneStruct_exref) {
          pcVar13 = (code *)0x0;
        }
        else {
LAB_1802a6712:
          pcVar13 = pcVar14;
          if ((pcVar14 != (code *)0x0) && (*(int *)pcVar14 + 1 != 0)) {
            *(int *)pcVar14 = *(int *)pcVar14 + 1;
          }
        }
      }
      else {
        *(int **)(pcVar14 + 0x10) = piVar2;
        if (*piVar2 + 1 != 0) {
          *piVar2 = *piVar2 + 1;
        }
        plVar8 = (longlong *)plVar9[5];
        if (pcVar14 != _Py_NoneStruct_exref) goto LAB_1802a6712;
        pcVar13 = (code *)0x0;
      }
      plVar9[5] = (longlong)pcVar13;
      if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))();
      }
      if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
         (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
        (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
      }
      plVar8 = *(longlong **)(param_1 + 0xf8);
      *(longlong **)(param_1 + 0xf8) = plVar9;
      if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))();
      }
      iVar6 = FUN_18000ce90(param_1,*(undefined8 *)(*(longlong *)(param_1 + 0xf8) + 8),
                            *(undefined8 *)PyExc_BaseException_exref);
      plVar10 = local_80;
      if (iVar6 == 0) {
        cVar5 = FUN_180373290(param_1,&local_90);
        uVar15 = 0x7a;
        if (cVar5 != '\0') {
          uVar15 = 0;
        }
        lVar11 = local_90[5];
        plVar12 = local_90;
        plVar10 = local_80;
        if ((lVar11 != 0) && (*(longlong **)(lVar11 + 0x18) == plVar4)) {
          *(undefined4 *)(plVar4 + 5) = *(undefined4 *)(lVar11 + 0x24);
        }
      }
      else {
        local_50 = *(code **)(param_1 + 0xf8);
        local_70 = 0;
        local_58 = *(longlong **)(local_50 + 8);
        local_48 = *(code **)(local_50 + 0x28);
        if (*(code **)(local_50 + 0x28) == (code *)0x0) {
          local_48 = _Py_NoneStruct_exref;
        }
        *(undefined4 *)(plVar4 + 5) = 0x7a;
        plVar8 = (longlong *)FUN_180378890(param_1,local_80,&local_58);
        if (plVar8 != (longlong *)0x0) {
          iVar6 = FUN_180001510(plVar8);
          if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
            (**(code **)(plVar8[1] + 0x30))(plVar8);
          }
          if (iVar6 != -1) {
            if (iVar6 != 0) {
              plVar8 = *(longlong **)(param_1 + 0xf8);
              *(int **)(param_1 + 0xf8) = piVar3;
              if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))();
              }
              goto LAB_1802a68ae;
            }
            cVar5 = FUN_180373290(param_1,&local_90);
            uVar15 = 0x7a;
            if (cVar5 != '\0') {
              uVar15 = 0;
            }
            lVar11 = local_90[5];
            plVar12 = local_90;
            if ((lVar11 != 0) && (*(longlong **)(lVar11 + 0x18) == plVar4)) {
              *(undefined4 *)(plVar4 + 5) = *(undefined4 *)(lVar11 + 0x24);
            }
            goto LAB_1802a69b2;
          }
        }
        plVar12 = *(longlong **)(param_1 + 0x60);
        uVar15 = 0x7a;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
LAB_1802a69b2:
      plVar7 = *(longlong **)(param_1 + 0xf8);
      *(int **)(param_1 + 0xf8) = piVar3;
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))();
      }
      plVar9 = local_68;
      plVar8 = local_78;
      if (local_70 != 0) {
        *(undefined4 *)(plVar4 + 5) = 0x7a;
/* novels.services.scheduler#47={"type": "tuple", "items": [{"tag": "n", "fileRange": [4367363, 4367364], "value": null}, {"tag": "n", "fileRange": [4367364, 4367365], "value": null}, {"tag": "n", "fileRange": [4367365, 4367366], "value": null}]} */
        plVar7 = (longlong *)FUN_1803793d0(param_1,plVar10,DAT_1803c3c08);
        plVar9 = local_68;
        plVar8 = local_78;
        if (plVar7 == (longlong *)0x0) {
          plVar7 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar12) && (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
            (**(code **)(plVar12[1] + 0x30))(plVar12);
            plVar9 = local_68;
            plVar8 = local_78;
          }
          goto LAB_1802a6a9c;
        }
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
          plVar9 = local_68;
          plVar8 = local_78;
        }
      }
    }
    if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
      (**(code **)(local_88[1] + 0x30))();
    }
    if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
  }
LAB_1802a693e:
  piVar3 = (int *)plVar12[5];
  if (piVar3 == (int *)0x0) {
    pcVar14 = (code *)FUN_18038e2a0(plVar4,uVar15);
    plVar7 = (longlong *)plVar12[5];
    pcVar13 = pcVar14;
    if (pcVar14 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_1802a6b5c:
      pcVar14 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1802a6ba5;
    pcVar14 = (code *)FUN_18038e2a0(plVar4,uVar15);
    *(int **)(pcVar14 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar7 = (longlong *)plVar12[5];
    pcVar13 = pcVar14;
    if (pcVar14 != _Py_NoneStruct_exref) goto LAB_1802a6b5c;
    pcVar13 = (code *)0x0;
  }
  plVar12[5] = (longlong)pcVar13;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar14 != (code *)0x0) && (-1 < *(int *)pcVar14)) &&
     (*(longlong *)pcVar14 = *(longlong *)pcVar14 + -1, *(longlong *)pcVar14 == 0)) {
    (**(code **)(*(longlong *)(pcVar14 + 8) + 0x30))(pcVar14);
  }
LAB_1802a6ba5:
  FUN_18036d0a0(plVar4,&DAT_180394940,plVar1);
  plVar7 = DAT_1803c3a28;
  if (plVar4 == DAT_1803c3a28) {
    if ((-1 < (int)*DAT_1803c3a28) && (*DAT_1803c3a28 = *DAT_1803c3a28 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803c3a28 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar12;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
