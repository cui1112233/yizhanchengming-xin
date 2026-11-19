/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801cb170 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

int * FUN_1801cb170(longlong param_1,int *param_2,undefined8 param_3)

{
  int *piVar1;
  code *pcVar2;
  char cVar3;
  int iVar4;
  undefined8 *puVar5;
  undefined8 uVar6;
  longlong lVar7;
  undefined8 uVar8;
  longlong *plVar9;
  longlong lVar10;
  longlong *plVar11;
  longlong lVar12;
  longlong *plVar13;
  longlong *plVar14;
  undefined1 auStack_c8 [32];
  undefined8 local_a8;
  longlong local_98;
  code *local_90;
  longlong *local_88;
  undefined8 local_80;
  longlong *local_78;
  undefined8 local_70;
  undefined8 local_68;
  code *local_60;
  code *local_58;
  undefined8 local_50;
  ulonglong local_48;
  
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_c8;
  DAT_1803ca608 = *(undefined8 *)(param_2 + 4);
  DAT_1803ca610 = param_2;
  if (DAT_1803b2afd == '\0') {
    if (DAT_1803b2afc == '\0') {
/* novels.prompts.inkos.reviser_prompts#0="- " */
      FUN_180388a30(param_1,&DAT_1803bbe10,"novels.prompts.inkos.reviser_prompts");
      DAT_1803b2afc = '\x01';
    }
    FUN_1801c0d90();
    DAT_1803b2afd = '\x01';
  }
  FUN_1801c0980(DAT_1803ca608,DAT_1803ab330,DAT_1803ab560);
  puVar5 = (undefined8 *)FUN_1800015a0(DAT_1803ca608,DAT_1803ab2a8);
  uVar8 = 0;
  if (puVar5 != (undefined8 *)0x0) {
    uVar8 = *puVar5;
  }
  uVar6 = PyUnicode_GetLength(uVar8);
  local_a8 = CONCAT44(local_a8._4_4_,0xffffffff);
  lVar7 = PyUnicode_Find(uVar8,DAT_1803ab4a8,0,uVar6);
  if (lVar7 != -1) {
    uVar8 = PyUnicode_Substring(uVar8,0,lVar7);
    FUN_1801c0a10(DAT_1803ca608,DAT_1803ab2b0,uVar8);
  }
  plVar9 = (longlong *)FUN_1800015a0(DAT_1803ca608,DAT_1803ab300);
  if ((plVar9 == (longlong *)0x0) || (*plVar9 == 0)) {
    uVar8 = PyModule_GetDict(DAT_1803ab780);
    FUN_1801c0980(DAT_1803ca608,DAT_1803ab300,uVar8);
  }
  uVar8 = FUN_18038d790(param_3);
  FUN_1801c0980(DAT_1803ca608,DAT_1803ab328,uVar8);
  if (DAT_1803c9fd0 == 0) {
    DAT_1803c9fd0 = PyImport_ImportModule("importlib._bootstrap");
  }
  plVar9 = (longlong *)PyObject_GetAttrString(DAT_1803c9fd0,"_spec_from_module");
  lVar7 = FUN_180376770(param_1,plVar9,DAT_1803ca610);
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if (lVar7 == 0) {
    PyErr_PrintEx(0);
                    /* WARNING: Subroutine does not return */
    abort();
  }
  FUN_18037f3a0(param_1,lVar7,DAT_1803ab4d0,_Py_TrueStruct_exref);
  FUN_1801c0a10(DAT_1803ca608,DAT_1803ab4c8,lVar7);
  local_98 = DAT_180391dc8;
/* novels.prompts.inkos.reviser_prompts#117="__doc__" */
  FUN_1801c0980(DAT_1803ca608,DAT_1803bc1b8,_Py_NoneStruct_exref);
/* novels.prompts.inkos.reviser_prompts#118="__file__" */
  FUN_1801c0980(DAT_1803ca608,DAT_1803bc1c0,DAT_1803bbe00);
  lVar7 = FUN_18036ccf0(DAT_1803bbd08,DAT_1803ca610);
  FUN_180001d00(param_1,lVar7);
  uVar8 = DAT_1803bbe00;
  uVar6 = FUN_1801c0ca0(param_1);
/* novels.prompts.inkos.reviser_prompts#119="origin" */
  cVar3 = FUN_18037f3a0(param_1,uVar6,DAT_1803bc1c8,uVar8);
  pcVar2 = _Py_TrueStruct_exref;
  if (cVar3 == '\0') {
LAB_1801cb411:
    lVar10 = *(longlong *)(param_1 + 0x60);
    uVar8 = 1;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    uVar8 = FUN_1801c0ca0(param_1);
/* novels.prompts.inkos.reviser_prompts#120="has_location" */
    cVar3 = FUN_18037f3a0(param_1,uVar8,DAT_1803bc1d0,pcVar2);
    if (cVar3 == '\0') goto LAB_1801cb411;
/* novels.prompts.inkos.reviser_prompts#121="__cached__" */
    FUN_1801c0980(DAT_1803ca608,DAT_1803bc1d8,_Py_NoneStruct_exref);
    uVar8 = FUN_180370ec0(param_1);
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
    FUN_1801c0a10(DAT_1803ca608,DAT_1803bc1e0,uVar8);
    uVar8 = FUN_180372070();
/* novels.prompts.inkos.reviser_prompts#123="annotations" */
    uVar8 = FUN_18037f0c0(param_1,uVar8,DAT_1803bc1e8);
/* novels.prompts.inkos.reviser_prompts#123="annotations" */
    FUN_1801c0a10(DAT_1803ca608,DAT_1803bc1e8,uVar8);
    lVar10 = DAT_1803ab578;
    local_50 = DAT_1803ab250;
    local_60 = _Py_NoneStruct_exref;
/* novels.prompts.inkos.reviser_prompts#124="re" */
    local_70 = DAT_1803bc1f0;
    local_68 = DAT_1803ca608;
    *(undefined4 *)(lVar7 + 0x28) = 3;
    local_58 = local_60;
    if (lVar10 == 0) {
      lVar10 = FUN_18035f440("__import__");
      DAT_1803ab578 = lVar10;
    }
    lVar10 = FUN_18037a640(param_1,lVar10,&local_70);
    if (lVar10 == 0) {
      lVar10 = *(longlong *)(param_1 + 0x60);
      uVar8 = 3;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801ce29b;
    }
/* novels.prompts.inkos.reviser_prompts#124="re" */
    FUN_1801c0a10(DAT_1803ca608,DAT_1803bc1f0,lVar10);
    local_58 = DAT_1803bc200;
    lVar10 = DAT_1803ab578;
    local_50 = DAT_1803ab250;
    local_60 = _Py_NoneStruct_exref;
/* novels.prompts.inkos.reviser_prompts#125="dataclasses" */
    local_70 = DAT_1803bc1f8;
    local_68 = DAT_1803ca608;
    *(undefined4 *)(lVar7 + 0x28) = 4;
    if (lVar10 == 0) {
      lVar10 = FUN_18035f440("__import__");
      DAT_1803ab578 = lVar10;
    }
    plVar9 = (longlong *)FUN_18037a640(param_1,lVar10,&local_70);
    if (plVar9 == (longlong *)0x0) {
LAB_1801cb5c0:
      lVar10 = *(longlong *)(param_1 + 0x60);
      uVar8 = 4;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
      if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
        lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc208);
      }
      else {
        local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
        lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc208);
      }
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (lVar10 == 0) goto LAB_1801cb5c0;
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
      FUN_1801c0a10(DAT_1803ca608,DAT_1803bc208,lVar10);
      plVar9 = (longlong *)FUN_180372110();
      iVar4 = (int)*plVar9 + 1;
      if (iVar4 != 0) {
        *(int *)plVar9 = iVar4;
      }
      iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
      if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#128="Any" */
        lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc210);
      }
      else {
        local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#128="Any" */
        lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc210);
      }
      if (lVar10 == 0) {
LAB_1801cb774:
        lVar10 = *(longlong *)(param_1 + 0x60);
        uVar8 = 5;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
      }
      else {
/* novels.prompts.inkos.reviser_prompts#128="Any" */
        FUN_1801c0a10(DAT_1803ca608,DAT_1803bc210,lVar10);
        iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
        if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
          lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc218);
        }
        else {
          local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
          lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc218);
        }
        if (lVar10 == 0) goto LAB_1801cb774;
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
        FUN_1801c0a10(DAT_1803ca608,DAT_1803bc218,lVar10);
        iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
        if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#130="Optional" */
          lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc220);
        }
        else {
          local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#130="Optional" */
          lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc220);
        }
        if (lVar10 == 0) goto LAB_1801cb774;
/* novels.prompts.inkos.reviser_prompts#130="Optional" */
        FUN_1801c0a10(DAT_1803ca608,DAT_1803bc220,lVar10);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        local_58 = DAT_1803bc230;
        lVar10 = DAT_1803ab578;
        local_50 = DAT_1803ab258;
        local_60 = _Py_NoneStruct_exref;
/* novels.prompts.inkos.reviser_prompts#131="genre_profile_defaults" */
        local_70 = DAT_1803bc228;
        local_68 = DAT_1803ca608;
        *(undefined4 *)(lVar7 + 0x28) = 7;
        if (lVar10 == 0) {
          lVar10 = FUN_18035f440("__import__");
          DAT_1803ab578 = lVar10;
        }
        plVar9 = (longlong *)FUN_18037a640(param_1,lVar10,&local_70);
        if (plVar9 == (longlong *)0x0) {
LAB_1801cb845:
          lVar10 = *(longlong *)(param_1 + 0x60);
          uVar8 = 7;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
          if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#133="GenreProfile" */
            lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc238);
          }
          else {
            local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#133="GenreProfile" */
            lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc238);
          }
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (lVar10 == 0) goto LAB_1801cb845;
/* novels.prompts.inkos.reviser_prompts#133="GenreProfile" */
          FUN_1801c0a10(DAT_1803ca608,DAT_1803bc238,lVar10);
          local_58 = DAT_1803bc248;
          lVar10 = DAT_1803ab578;
          local_50 = DAT_1803ab258;
          local_60 = _Py_NoneStruct_exref;
/* novels.prompts.inkos.reviser_prompts#134="length_governance" */
          local_70 = DAT_1803bc240;
          local_68 = DAT_1803ca608;
          *(undefined4 *)(lVar7 + 0x28) = 8;
          if (lVar10 == 0) {
            lVar10 = FUN_18035f440("__import__");
            DAT_1803ab578 = lVar10;
          }
          plVar9 = (longlong *)FUN_18037a640(param_1,lVar10,&local_70);
          if (plVar9 == (longlong *)0x0) {
LAB_1801cb94c:
            lVar10 = *(longlong *)(param_1 + 0x60);
            uVar8 = 8;
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            iVar4 = FUN_18000cc50(plVar9,PyModule_Type_exref);
            if (iVar4 == 0) {
/* novels.prompts.inkos.reviser_prompts#136="LengthSpec" */
              lVar10 = FUN_180371890(param_1,plVar9,DAT_1803bc250);
            }
            else {
              local_a8 = DAT_1803ab250;
/* novels.prompts.inkos.reviser_prompts#136="LengthSpec" */
              lVar10 = FUN_180371d80(param_1,plVar9,DAT_1803ca608,DAT_1803bc250);
            }
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if (lVar10 == 0) goto LAB_1801cb94c;
/* novels.prompts.inkos.reviser_prompts#136="LengthSpec" */
            FUN_1801c0a10(DAT_1803ca608,DAT_1803bc250,lVar10);
            lVar10 = FUN_1801c0aa0(param_1);
            if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
              FUN_180373340(param_1,&local_98,DAT_1803bc218);
              uVar8 = 10;
              lVar10 = local_98;
            }
            else {
              lVar10 = FUN_180001b80(param_1,lVar10,DAT_1803bc258);
              if (lVar10 == 0) {
                lVar10 = *(longlong *)(param_1 + 0x60);
                uVar8 = 10;
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
/* novels.prompts.inkos.reviser_prompts#138="ReviseMode" */
                FUN_1801c0a10(DAT_1803ca608,DAT_1803bc260,lVar10);
                lVar10 = FUN_1801c0aa0(param_1);
                if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
                  FUN_180373340(param_1,&local_98,DAT_1803bc218);
                  uVar8 = 0xb;
                  lVar10 = local_98;
                }
                else {
                  lVar10 = FUN_180001b80(param_1,lVar10,DAT_1803bc268);
                  if (lVar10 == 0) {
                    lVar10 = *(longlong *)(param_1 + 0x60);
                    uVar8 = 0xb;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
/* novels.prompts.inkos.reviser_prompts#140="AutoOutputMode" */
                    FUN_1801c0a10(DAT_1803ca608,DAT_1803bc270,lVar10);
                    lVar10 = FUN_1801c0aa0(param_1);
                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#129="Literal" */
                      FUN_180373340(param_1,&local_98,DAT_1803bc218);
                      uVar8 = 0xc;
                      lVar10 = local_98;
                    }
                    else {
                      lVar10 = FUN_180001b80(param_1,lVar10,DAT_1803bc278);
                      if (lVar10 == 0) {
                        lVar10 = *(longlong *)(param_1 + 0x60);
                        uVar8 = 0xc;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
/* novels.prompts.inkos.reviser_prompts#142="Language" */
                        FUN_1801c0a10(DAT_1803ca608,DAT_1803bc280,lVar10);
                        uVar8 = FUN_18036fb80(param_1,DAT_1803bc288);
/* novels.prompts.inkos.reviser_prompts#87="MODE_DESCRIPTIONS" */
                        FUN_1801c0a10(DAT_1803ca608,DAT_1803bc0c8,uVar8);
/* novels.prompts.inkos.reviser_prompts#144="dict[str, str]" */
                        uVar8 = DAT_1803bc290;
                        lVar10 = FUN_1801c0c00(param_1);
                        if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                          FUN_180373340(param_1,&local_98,DAT_1803bc1e0);
                          uVar8 = 0xe;
                          lVar10 = local_98;
                        }
                        else {
/* novels.prompts.inkos.reviser_prompts#87="MODE_DESCRIPTIONS" */
                          cVar3 = FUN_18000d470(param_1,lVar10,DAT_1803bc0c8,uVar8);
                          if (cVar3 == '\0') {
                            lVar10 = *(longlong *)(param_1 + 0x60);
                            uVar8 = 0xe;
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
                            lVar10 = FUN_1801c0d40(param_1);
                            if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                              FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                              uVar8 = 0x23;
                              lVar10 = local_98;
                            }
                            else {
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                              plVar9 = (longlong *)FUN_18037f0c0(param_1,lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#146="Paragraph uniformity|段落等长" */
                              uVar8 = DAT_1803bc2a0;
                              if (plVar9 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_98);
                                uVar8 = 0x23;
                                lVar10 = local_98;
                              }
                              else {
                                lVar10 = FUN_1801c0d40(param_1);
                                if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                  FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
LAB_1801cbc31:
                                  uVar8 = 0x23;
                                  lVar10 = local_98;
                                }
                                else {
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                  plVar11 = (longlong *)FUN_18037f0c0(param_1,lVar10,DAT_1803bc2a8);
                                  if (plVar11 == (longlong *)0x0) {
                                    FUN_18000ce30(param_1,&local_98);
                                    if (((int)*plVar9 < 0) || (*plVar9 = *plVar9 + -1, *plVar9 != 0)
                                       ) goto LAB_1801cbc31;
                                    lVar10 = FUN_18000cc40(plVar9);
                                    (**(code **)(lVar10 + 0x30))(plVar9);
                                    uVar8 = 0x23;
                                    lVar10 = local_98;
                                  }
                                  else {
                                    *(undefined4 *)(lVar7 + 0x28) = 0x23;
                                    local_80 = uVar8;
                                    local_78 = plVar11;
                                    lVar10 = FUN_180377750(param_1,plVar9,&local_80);
                                    if ((-1 < (int)*plVar9) &&
                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                      lVar12 = FUN_18000cc40(plVar9);
                                      (**(code **)(lVar12 + 0x30))(plVar9);
                                    }
                                    if ((-1 < (int)*plVar11) &&
                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                      lVar12 = FUN_18000cc40(plVar11);
                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                    }
                                    if (lVar10 == 0) {
                                      FUN_18000ce30(param_1,&local_98);
                                      uVar8 = 0x23;
                                      lVar10 = local_98;
                                    }
                                    else {
                                      plVar9 = (longlong *)FUN_180372160(param_1,9);
                                      FUN_180024750(plVar9,0,lVar10);
                                      lVar10 = FUN_1801c0d40(param_1);
                                      if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                        FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                                        uVar8 = 0x24;
                                      }
                                      else {
                                        plVar11 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  FUN_18037f0c0(param_1,lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#148="Hedge density|套话密度" */
                                        uVar8 = DAT_1803bc2b0;
                                        if (plVar11 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_98);
                                          uVar8 = 0x24;
                                        }
                                        else {
                                          lVar10 = FUN_1801c0d40(param_1);
                                          if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                            FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                                          }
                                          else {
                                            plVar13 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                      FUN_18037f0c0(param_1,lVar10,DAT_1803bc2a8);
                                            if (plVar13 != (longlong *)0x0) {
                                              *(undefined4 *)(lVar7 + 0x28) = 0x24;
                                              local_80 = uVar8;
                                              local_78 = plVar13;
                                              lVar10 = FUN_180377750(param_1,plVar11,&local_80);
                                              if ((-1 < (int)*plVar11) &&
                                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                lVar12 = FUN_18000cc40(plVar11);
                                                (**(code **)(lVar12 + 0x30))(plVar11);
                                              }
                                              if ((-1 < (int)*plVar13) &&
                                                 (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                lVar12 = FUN_18000cc40(plVar13);
                                                (**(code **)(lVar12 + 0x30))(plVar13);
                                              }
                                              if (lVar10 == 0) {
                                                FUN_18000ce30(param_1,&local_98);
                                                uVar8 = 0x24;
                                              }
                                              else {
                                                FUN_180024750(plVar9,1,lVar10);
                                                lVar10 = FUN_1801c0d40(param_1);
                                                if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                  FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                                                  uVar8 = 0x25;
                                                }
                                                else {
                                                  plVar11 = (longlong *)
                                                            FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                          DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#149="Formulaic transitions|公式化转折" */
                                                  uVar8 = DAT_1803bc2b8;
                                                  if (plVar11 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x25;
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x25;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x25;
                                                        }
                                                        else {
                                                          FUN_180024750(plVar9,2,lVar10);
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                            uVar8 = 0x26;
                                                          }
                                                          else {
                                                            plVar11 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                    DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#150="List-like structure|列表式结构" */
                                                            uVar8 = DAT_1803bc2c0;
                                                            if (plVar11 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_98);
                                                              uVar8 = 0x26;
                                                            }
                                                            else {
                                                              lVar10 = FUN_1801c0d40(param_1);
                                                              if (lVar10 == 0) {
                                                                FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                              DAT_1803bc1f0);
                                                              }
                                                              else {
                                                                plVar13 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                  lVar10,DAT_1803bc2a8);
                                                  if (plVar13 != (longlong *)0x0) {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x26;
                                                    local_80 = uVar8;
                                                    local_78 = plVar13;
                                                    lVar10 = FUN_180377750(param_1,plVar11,&local_80
                                                                          );
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar12 + 0x30))(plVar13);
                                                    }
                                                    if (lVar10 == 0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x26;
                                                    }
                                                    else {
                                                      FUN_180024750(plVar9,3,lVar10);
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                        uVar8 = 0x27;
                                                      }
                                                      else {
                                                        plVar11 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#151="Cross-chapter repetition|跨章重复" */
                                                        uVar8 = DAT_1803bc2c8;
                                                        if (plVar11 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x27;
                                                        }
                                                        else {
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                    DAT_1803bc2a8);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              *(undefined4 *)(lVar7 + 0x28) = 0x27;
                                                              local_80 = uVar8;
                                                              local_78 = plVar13;
                                                              lVar10 = FUN_180377750(param_1,plVar11
                                                                                     ,&local_80);
                                                              if ((-1 < (int)*plVar11) &&
                                                                 (*plVar11 = *plVar11 + -1,
                                                                 *plVar11 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar11);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar11);
                                                              }
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              if (lVar10 == 0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x27;
                                                              }
                                                              else {
                                                                FUN_180024750(plVar9,4,lVar10);
                                                                lVar10 = FUN_1801c0d40(param_1);
                                                                if (lVar10 == 0) {
                                                                  FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                                DAT_1803bc1f0);
                                                                  uVar8 = 0x28;
                                                                }
                                                                else {
                                                                  plVar11 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#152="AI-tell word density" */
                                                  uVar8 = DAT_1803bc2d0;
                                                  if (plVar11 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x28;
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x28;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x28;
                                                        }
                                                        else {
                                                          FUN_180024750(plVar9,5,lVar10);
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                            uVar8 = 0x29;
                                                          }
                                                          else {
                                                            plVar11 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                    DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#153="Fatigue word|高疲劳词" */
                                                            uVar8 = DAT_1803bc2d8;
                                                            if (plVar11 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_98);
                                                              uVar8 = 0x29;
                                                            }
                                                            else {
                                                              lVar10 = FUN_1801c0d40(param_1);
                                                              if (lVar10 == 0) {
                                                                FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                              DAT_1803bc1f0);
                                                              }
                                                              else {
                                                                plVar13 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                  lVar10,DAT_1803bc2a8);
                                                  if (plVar13 != (longlong *)0x0) {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x29;
                                                    local_80 = uVar8;
                                                    local_78 = plVar13;
                                                    lVar10 = FUN_180377750(param_1,plVar11,&local_80
                                                                          );
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar12 + 0x30))(plVar13);
                                                    }
                                                    if (lVar10 == 0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x29;
                                                    }
                                                    else {
                                                      FUN_180024750(plVar9,6,lVar10);
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                        uVar8 = 0x2a;
                                                      }
                                                      else {
                                                        plVar11 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#154="Information Boundary Check|信息越界" */
                                                        uVar8 = DAT_1803bc2e0;
                                                        if (plVar11 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x2a;
                                                        }
                                                        else {
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                    DAT_1803bc2a8);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              *(undefined4 *)(lVar7 + 0x28) = 0x2a;
                                                              local_80 = uVar8;
                                                              local_78 = plVar13;
                                                              lVar10 = FUN_180377750(param_1,plVar11
                                                                                     ,&local_80);
                                                              if ((-1 < (int)*plVar11) &&
                                                                 (*plVar11 = *plVar11 + -1,
                                                                 *plVar11 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar11);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar11);
                                                              }
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              if (lVar10 == 0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x2a;
                                                                goto LAB_1801cc71e;
                                                              }
                                                              FUN_180024750(plVar9,7,lVar10);
                                                              lVar10 = FUN_1801c0d40(param_1);
                                                              if (lVar10 == 0) {
                                                                FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                              DAT_1803bc1f0);
                                                              }
                                                              else {
                                                                plVar11 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#155="Knowledge Base Pollution|知识库污染" */
                                                  uVar8 = DAT_1803bc2e8;
                                                  if (plVar11 == (longlong *)0x0) {
LAB_1801cc710:
                                                    FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x2b;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) goto LAB_1801cc710;
                                                        FUN_180024750(plVar9,8,lVar10);
/* novels.prompts.inkos.reviser_prompts#40="LOCAL_ONLY_PATTERNS" */
                                                        FUN_1801c0a10(DAT_1803ca608,DAT_1803bbf50,
                                                                      plVar9);
                                                        lVar10 = FUN_1801c0d40(param_1);
                                                        if (lVar10 == 0) {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                        DAT_1803bc1f0);
                                                          uVar8 = 0x2f;
                                                          lVar10 = local_98;
                                                        }
                                                        else {
                                                          plVar9 = (longlong *)
                                                                   FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                 DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#156="OOC|人设|Character Fidelity|Character Matrix|Character.*Consistency" */
                                                          uVar8 = DAT_1803bc2f0;
                                                          if (plVar9 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_98);
                                                            uVar8 = 0x2f;
                                                            lVar10 = local_98;
                                                          }
                                                          else {
                                                            lVar10 = FUN_1801c0d40(param_1);
                                                            if (lVar10 == 0) {
                                                              FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                            DAT_1803bc1f0);
                                                            }
                                                            else {
                                                              plVar11 = (longlong *)
                                                                        FUN_18037f0c0(param_1,lVar10
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                      ,DAT_1803bc2a8
                                                                                     );
                                                              if (plVar11 != (longlong *)0x0) {
                                                                *(undefined4 *)(lVar7 + 0x28) = 0x2f
                                                                ;
                                                                local_80 = uVar8;
                                                                local_78 = plVar11;
                                                                lVar10 = FUN_180377750(param_1,
                                                  plVar9,&local_80);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar12 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar12 + 0x30))(plVar11);
                                                  }
                                                  if (lVar10 == 0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x2f;
                                                    lVar10 = local_98;
                                                    goto LAB_1801ce29b;
                                                  }
                                                  plVar9 = (longlong *)FUN_180372160(param_1,0xc);
                                                  FUN_180024750(plVar9,0,lVar10);
                                                  lVar10 = FUN_1801c0d40(param_1);
                                                  if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                    FUN_180373340(param_1,&local_98,DAT_1803bc1f0);
                                                    uVar8 = 0x30;
                                                  }
                                                  else {
                                                    plVar11 = (longlong *)
                                                              FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                            DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#157="Mainline.*Drift|主线偏离|Outline Drift|大纲偏离|Chapter Memo Drift|章节备忘偏离" */
                                                    uVar8 = DAT_1803bc2f8;
                                                    if (plVar11 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x30;
                                                    }
                                                    else {
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                      }
                                                      else {
                                                        plVar13 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                DAT_1803bc2a8);
                                                        if (plVar13 != (longlong *)0x0) {
                                                          *(undefined4 *)(lVar7 + 0x28) = 0x30;
                                                          local_80 = uVar8;
                                                          local_78 = plVar13;
                                                          lVar10 = FUN_180377750(param_1,plVar11,
                                                                                 &local_80);
                                                          if ((-1 < (int)*plVar11) &&
                                                             (*plVar11 = *plVar11 + -1,
                                                             *plVar11 == 0)) {
                                                            lVar12 = FUN_18000cc40(plVar11);
                                                            (**(code **)(lVar12 + 0x30))(plVar11);
                                                          }
                                                          if ((-1 < (int)*plVar13) &&
                                                             (*plVar13 = *plVar13 + -1,
                                                             *plVar13 == 0)) {
                                                            lVar12 = FUN_18000cc40(plVar13);
                                                            (**(code **)(lVar12 + 0x30))(plVar13);
                                                          }
                                                          if (lVar10 == 0) {
                                                            FUN_18000ce30(param_1,&local_98);
                                                            uVar8 = 0x30;
                                                          }
                                                          else {
                                                            FUN_180024750(plVar9,1,lVar10);
                                                            lVar10 = FUN_1801c0d40(param_1);
                                                            if (lVar10 == 0) {
                                                              FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                            DAT_1803bc1f0);
                                                              uVar8 = 0x31;
                                                            }
                                                            else {
                                                              plVar11 = (longlong *)
                                                                        FUN_18037f0c0(param_1,lVar10
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                      ,DAT_1803bc298
                                                                                     );
/* novels.prompts.inkos.reviser_prompts#158="Conflict|冲突乏力|Payoff Dilution|爽点虚化" */
                                                              uVar8 = DAT_1803bc300;
                                                              if (plVar11 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x31;
                                                              }
                                                              else {
                                                                lVar10 = FUN_1801c0d40(param_1);
                                                                if (lVar10 == 0) {
                                                                  FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                                DAT_1803bc1f0);
                                                                }
                                                                else {
                                                                  plVar13 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                  lVar10,DAT_1803bc2a8);
                                                  if (plVar13 != (longlong *)0x0) {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x31;
                                                    local_80 = uVar8;
                                                    local_78 = plVar13;
                                                    lVar10 = FUN_180377750(param_1,plVar11,&local_80
                                                                          );
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar12 + 0x30))(plVar13);
                                                    }
                                                    if (lVar10 == 0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x31;
                                                    }
                                                    else {
                                                      FUN_180024750(plVar9,2,lVar10);
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                        uVar8 = 0x32;
                                                      }
                                                      else {
                                                        plVar11 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#159="Timeline|时间线" */
                                                        uVar8 = DAT_1803bc308;
                                                        if (plVar11 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x32;
                                                        }
                                                        else {
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                    DAT_1803bc2a8);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              *(undefined4 *)(lVar7 + 0x28) = 0x32;
                                                              local_80 = uVar8;
                                                              local_78 = plVar13;
                                                              lVar10 = FUN_180377750(param_1,plVar11
                                                                                     ,&local_80);
                                                              if ((-1 < (int)*plVar11) &&
                                                                 (*plVar11 = *plVar11 + -1,
                                                                 *plVar11 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar11);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar11);
                                                              }
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              if (lVar10 == 0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x32;
                                                              }
                                                              else {
                                                                FUN_180024750(plVar9,3,lVar10);
                                                                lVar10 = FUN_1801c0d40(param_1);
                                                                if (lVar10 == 0) {
                                                                  FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                                DAT_1803bc1f0);
                                                                  uVar8 = 0x33;
                                                                }
                                                                else {
                                                                  plVar11 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#160="Hook Check|伏笔检查|Hook.*Debt|伏笔.*债|未兑现" */
                                                  uVar8 = DAT_1803bc310;
                                                  if (plVar11 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x33;
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x33;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x33;
                                                        }
                                                        else {
                                                          FUN_180024750(plVar9,4,lVar10);
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                            uVar8 = 0x34;
                                                          }
                                                          else {
                                                            plVar11 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                    DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#161="Power Scaling|战力崩坏|金手指" */
                                                            uVar8 = DAT_1803bc318;
                                                            if (plVar11 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_98);
                                                              uVar8 = 0x34;
                                                            }
                                                            else {
                                                              lVar10 = FUN_1801c0d40(param_1);
                                                              if (lVar10 == 0) {
                                                                FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                              DAT_1803bc1f0);
                                                              }
                                                              else {
                                                                plVar13 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                  lVar10,DAT_1803bc2a8);
                                                  if (plVar13 != (longlong *)0x0) {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x34;
                                                    local_80 = uVar8;
                                                    local_78 = plVar13;
                                                    lVar10 = FUN_180377750(param_1,plVar11,&local_80
                                                                          );
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar12 + 0x30))(plVar13);
                                                    }
                                                    if (lVar10 == 0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x34;
                                                    }
                                                    else {
                                                      FUN_180024750(plVar9,5,lVar10);
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                        uVar8 = 0x35;
                                                      }
                                                      else {
                                                        plVar11 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#162="Pacing|节奏" */
                                                        uVar8 = DAT_1803bc320;
                                                        if (plVar11 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x35;
                                                        }
                                                        else {
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                    DAT_1803bc2a8);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              *(undefined4 *)(lVar7 + 0x28) = 0x35;
                                                              local_80 = uVar8;
                                                              local_78 = plVar13;
                                                              lVar10 = FUN_180377750(param_1,plVar11
                                                                                     ,&local_80);
                                                              if ((-1 < (int)*plVar11) &&
                                                                 (*plVar11 = *plVar11 + -1,
                                                                 *plVar11 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar11);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar11);
                                                              }
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              if (lVar10 == 0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x35;
                                                              }
                                                              else {
                                                                FUN_180024750(plVar9,6,lVar10);
                                                                lVar10 = FUN_1801c0d40(param_1);
                                                                if (lVar10 == 0) {
                                                                  FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                                DAT_1803bc1f0);
                                                                  uVar8 = 0x36;
                                                                }
                                                                else {
                                                                  plVar11 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#163="POV Consistency|视角" */
                                                  uVar8 = DAT_1803bc328;
                                                  if (plVar11 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x36;
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x36;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x36;
                                                        }
                                                        else {
                                                          FUN_180024750(plVar9,7,lVar10);
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                            uVar8 = 0x37;
                                                          }
                                                          else {
                                                            plVar11 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                    DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#164="Subplot Stagnation|支线停滞|Arc Flatline|弧线平坦" */
                                                            uVar8 = DAT_1803bc330;
                                                            if (plVar11 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_98);
                                                              uVar8 = 0x37;
                                                            }
                                                            else {
                                                              lVar10 = FUN_1801c0d40(param_1);
                                                              if (lVar10 == 0) {
                                                                FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                              DAT_1803bc1f0);
                                                              }
                                                              else {
                                                                plVar13 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                  lVar10,DAT_1803bc2a8);
                                                  if (plVar13 != (longlong *)0x0) {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x37;
                                                    local_80 = uVar8;
                                                    local_78 = plVar13;
                                                    lVar10 = FUN_180377750(param_1,plVar11,&local_80
                                                                          );
                                                    if ((-1 < (int)*plVar11) &&
                                                       (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar11);
                                                      (**(code **)(lVar12 + 0x30))(plVar11);
                                                    }
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar12 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar12 + 0x30))(plVar13);
                                                    }
                                                    if (lVar10 == 0) {
                                                      FUN_18000ce30(param_1,&local_98);
                                                      uVar8 = 0x37;
                                                    }
                                                    else {
                                                      FUN_180024750(plVar9,8,lVar10);
                                                      lVar10 = FUN_1801c0d40(param_1);
                                                      if (lVar10 == 0) {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                      DAT_1803bc1f0);
                                                        uVar8 = 0x38;
                                                      }
                                                      else {
                                                        plVar11 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#165="Relationship Dynamics|关系动态|情感表达" */
                                                        uVar8 = DAT_1803bc338;
                                                        if (plVar11 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x38;
                                                        }
                                                        else {
                                                          lVar10 = FUN_1801c0d40(param_1);
                                                          if (lVar10 == 0) {
                                                            FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                          DAT_1803bc1f0);
                                                          }
                                                          else {
                                                            plVar13 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                    DAT_1803bc2a8);
                                                            if (plVar13 != (longlong *)0x0) {
                                                              *(undefined4 *)(lVar7 + 0x28) = 0x38;
                                                              local_80 = uVar8;
                                                              local_78 = plVar13;
                                                              lVar10 = FUN_180377750(param_1,plVar11
                                                                                     ,&local_80);
                                                              if ((-1 < (int)*plVar11) &&
                                                                 (*plVar11 = *plVar11 + -1,
                                                                 *plVar11 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar11);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar11);
                                                              }
                                                              if ((-1 < (int)*plVar13) &&
                                                                 (*plVar13 = *plVar13 + -1,
                                                                 *plVar13 == 0)) {
                                                                lVar12 = FUN_18000cc40(plVar13);
                                                                (**(code **)(lVar12 + 0x30))
                                                                          (plVar13);
                                                              }
                                                              if (lVar10 == 0) {
                                                                FUN_18000ce30(param_1,&local_98);
                                                                uVar8 = 0x38;
                                                              }
                                                              else {
                                                                FUN_180024750(plVar9,9,lVar10);
                                                                lVar10 = FUN_1801c0d40(param_1);
                                                                if (lVar10 == 0) {
                                                                  FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                                DAT_1803bc1f0);
                                                                  uVar8 = 0x39;
                                                                }
                                                                else {
                                                                  plVar11 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                  lVar10,DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#166="Incentive Chain|利益链" */
                                                  uVar8 = DAT_1803bc340;
                                                  if (plVar11 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_98);
                                                    uVar8 = 0x39;
                                                  }
                                                  else {
                                                    lVar10 = FUN_1801c0d40(param_1);
                                                    if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                      FUN_180373340(param_1,&local_98,DAT_1803bc1f0)
                                                      ;
                                                    }
                                                    else {
                                                      plVar13 = (longlong *)
                                                                FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                              DAT_1803bc2a8);
                                                      if (plVar13 != (longlong *)0x0) {
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x39;
                                                        local_80 = uVar8;
                                                        local_78 = plVar13;
                                                        lVar10 = FUN_180377750(param_1,plVar11,
                                                                               &local_80);
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar12 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if (lVar10 == 0) {
                                                          FUN_18000ce30(param_1,&local_98);
                                                          uVar8 = 0x39;
                                                          goto LAB_1801cd698;
                                                        }
                                                        FUN_180024750(plVar9,10,lVar10);
                                                        lVar10 = FUN_1801c0d40(param_1);
                                                        if (lVar10 == 0) {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                        DAT_1803bc1f0);
                                                        }
                                                        else {
                                                          plVar11 = (longlong *)
                                                                    FUN_18037f0c0(param_1,lVar10,
/* novels.prompts.inkos.reviser_prompts#145="compile" */
                                                                                  DAT_1803bc298);
/* novels.prompts.inkos.reviser_prompts#167="Canon Event|正典|Mainline Canon" */
                                                          uVar8 = DAT_1803bc348;
                                                          if (plVar11 == (longlong *)0x0) {
LAB_1801cd68a:
                                                            FUN_18000ce30(param_1,&local_98);
                                                          }
                                                          else {
                                                            lVar10 = FUN_1801c0d40(param_1);
                                                            if (lVar10 == 0) {
                                                              FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#124="re" */
                                                                            DAT_1803bc1f0);
                                                            }
                                                            else {
                                                              plVar13 = (longlong *)
                                                                        FUN_18037f0c0(param_1,lVar10
/* novels.prompts.inkos.reviser_prompts#147="I" */
                                                                                      ,DAT_1803bc2a8
                                                                                     );
                                                              if (plVar13 != (longlong *)0x0) {
                                                                *(undefined4 *)(lVar7 + 0x28) = 0x3a
                                                                ;
                                                                local_80 = uVar8;
                                                                local_78 = plVar13;
                                                                lVar10 = FUN_180377750(param_1,
                                                  plVar11,&local_80);
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar12 + 0x30))(plVar11);
                                                  }
                                                  if ((-1 < (int)*plVar13) &&
                                                     (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar13);
                                                    (**(code **)(lVar12 + 0x30))(plVar13);
                                                  }
                                                  if (lVar10 == 0) goto LAB_1801cd68a;
                                                  FUN_180024750(plVar9,0xb,lVar10);
/* novels.prompts.inkos.reviser_prompts#36="STRUCTURAL_PATTERNS" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bbf30,plVar9);
                                                  piVar1 = DAT_1803bc350;
                                                  uVar8 = FUN_18036fb80(param_1,DAT_1803bc358);
                                                  if (*piVar1 + 1 != 0) {
                                                    *piVar1 = *piVar1 + 1;
                                                  }
                                                  uVar8 = FUN_1801caf90(param_1,piVar1,uVar8);
/* novels.prompts.inkos.reviser_prompts#170="build_tiered_issue_list" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc360,uVar8);
                                                  uVar8 = FUN_18036fb80(param_1,DAT_1803bc368);
                                                  uVar8 = FUN_1801cafe0(param_1,uVar8);
/* novels.prompts.inkos.reviser_prompts#172="resolve_auto_output_mode" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc370,uVar8);
                                                  piVar1 = DAT_1803bc378;
                                                  uVar8 = FUN_18036fb80(param_1,DAT_1803bc380);
                                                  if (*piVar1 + 1 != 0) {
                                                    *piVar1 = *piVar1 + 1;
                                                  }
                                                  uVar8 = FUN_1801cb030(param_1,piVar1,uVar8);
/* novels.prompts.inkos.reviser_prompts#175="build_protagonist_block" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc388,uVar8);
                                                  uVar8 = FUN_18036fb80(param_1,DAT_1803bc390);
                                                  uVar6 = FUN_18036fb80(param_1,DAT_1803bc398);
                                                  uVar8 = FUN_1801cb080(param_1,uVar8,uVar6);
/* novels.prompts.inkos.reviser_prompts#178="build_auto_system_prompt" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc3a0,uVar8);
                                                  uVar8 = FUN_18036fb80(param_1,DAT_1803bc3a8);
                                                  uVar6 = FUN_18036fb80(param_1,DAT_1803bc3b0);
                                                  uVar8 = FUN_1801cb0d0(param_1,uVar8,uVar6);
/* novels.prompts.inkos.reviser_prompts#181="build_legacy_system_prompt" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc3b8,uVar8);
                                                  plVar9 = (longlong *)FUN_180370ec0(param_1);
                                                  plVar11 = (longlong *)FUN_180370ec0(param_1);
                                                  lVar10 = FUN_1801c0cf0(param_1);
                                                  if (lVar10 == 0) {
/* novels.prompts.inkos.reviser_prompts#127="dataclass" */
                                                    FUN_180373340(param_1,&local_98,DAT_1803bc208);
LAB_1801ce249:
                                                    uVar8 = 0x166;
                                                  }
                                                  else {
                                                    *(undefined4 *)(lVar7 + 0x28) = 0x166;
                                                    local_88 = (longlong *)
                                                               FUN_18037ccb0(param_1,lVar10,
                                                                             DAT_1803bc3c0 + 0x18,
                                                                             DAT_1803bc3c8);
                                                    if (local_88 == (longlong *)0x0) {
LAB_1801ce23d:
                                                      FUN_18000ce30(param_1,&local_98);
                                                      goto LAB_1801ce249;
                                                    }
                                                    iVar4 = (int)*plVar11 + 1;
                                                    if (iVar4 != 0) {
                                                      *(int *)plVar11 = iVar4;
                                                    }
                                                    uVar8 = FUN_1801c0c50(param_1);
/* novels.prompts.inkos.reviser_prompts#184="__module__" */
                                                    FUN_18009b380(plVar11,DAT_1803bc3d0,uVar8);
/* novels.prompts.inkos.reviser_prompts#186="__qualname__" */
                                                    FUN_18009b380(plVar11,DAT_1803bc3e0,
/* novels.prompts.inkos.reviser_prompts#185="ReviserUserPromptParams" */
                                                                  DAT_1803bc3d8);
                                                    plVar13 = (longlong *)FUN_180370ec0(param_1);
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                    FUN_18009b380(plVar11,DAT_1803bc1e0,plVar13);
                                                    if ((-1 < (int)*plVar13) &&
                                                       (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                                      lVar10 = FUN_18000cc40(plVar13);
                                                      (**(code **)(lVar10 + 0x30))(plVar13);
                                                    }
                                                    local_a8 = 8;
                                                    lVar10 = FUN_18036cd80(param_1,DAT_1803bbdd0,
                                                                           DAT_1803ca610,0);
                                                    FUN_180001d00(param_1,lVar10);
/* novels.prompts.inkos.reviser_prompts#187="int" */
                                                    uVar8 = DAT_1803bc3e8;
                                                    lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                           DAT_1803bc1e0);
                                                    if (lVar12 == 0) {
                                                      cVar3 = FUN_18009b370();
                                                      if (cVar3 == '\0') {
LAB_1801cd996:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x168;
                                                      }
                                                      else {
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                      DAT_1803bc1e0);
                                                        uVar8 = 0x168;
                                                      }
LAB_1801ce03f:
                                                      lVar12 = FUN_18000ce80(&local_98);
                                                      if (lVar12 == 0) {
                                                        uVar8 = FUN_18038e2a0(lVar10,uVar8);
LAB_1801ce06d:
                                                        FUN_180001300(&local_98,uVar8);
                                                      }
                                                      else if (*(longlong *)(lVar12 + 0x18) !=
                                                               lVar10) {
                                                        uVar8 = FUN_18000cc90(lVar12,lVar10,uVar8);
                                                        goto LAB_1801ce06d;
                                                      }
                                                      FUN_18036d0a0(lVar10,&DAT_180394940,0);
                                                      _guard_check_icall(lVar10);
                                                      FUN_180001d80(param_1);
                                                    }
                                                    else {
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#100="chapter_number" */
                                                                            DAT_1803bc130,uVar8);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      if (cVar3 == '\0') goto LAB_1801cd996;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x169;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cda06:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x169;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#102="issue_list" */
                                                                            DAT_1803bc140,uVar8);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      if (cVar3 == '\0') goto LAB_1801cda06;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16a;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cda76:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16a;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#104="current_state" */
                                                                            DAT_1803bc150,uVar8);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      if (cVar3 == '\0') goto LAB_1801cda76;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16b;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdae6:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16b;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#116="chapter_content" */
                                                                            DAT_1803bc1b0,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdae6;
/* novels.prompts.inkos.reviser_prompts#105="ledger_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc158,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16c;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdb6f:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16c;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#105="ledger_block" */
                                                                            DAT_1803bc158,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdb6f;
/* novels.prompts.inkos.reviser_prompts#106="hooks_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc160,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16d;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdbf8:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16d;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#106="hooks_block" */
                                                                            DAT_1803bc160,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdbf8;
/* novels.prompts.inkos.reviser_prompts#107="outline_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc168,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16e;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdc81:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16e;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#107="outline_block" */
                                                                            DAT_1803bc168,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdc81;
/* novels.prompts.inkos.reviser_prompts#108="bible_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc170,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x16f;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdd0a:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x16f;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#108="bible_block" */
                                                                            DAT_1803bc170,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdd0a;
/* novels.prompts.inkos.reviser_prompts#109="matrix_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc178,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x170;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdd93:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x170;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#109="matrix_block" */
                                                                            DAT_1803bc178,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdd93;
/* novels.prompts.inkos.reviser_prompts#110="summaries_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc180,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x171;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cde1c:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x171;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#110="summaries_block" */
                                                                            DAT_1803bc180,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cde1c;
/* novels.prompts.inkos.reviser_prompts#111="canon_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc188,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x172;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdea5:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x172;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#111="canon_block" */
                                                                            DAT_1803bc188,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdea5;
/* novels.prompts.inkos.reviser_prompts#112="fanfic_canon_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc190,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x173;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdf2e:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x173;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#112="fanfic_canon_block" */
                                                                            DAT_1803bc190,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdf2e;
/* novels.prompts.inkos.reviser_prompts#113="style_guide_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc198,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 != '\0') {
                                                          FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                        DAT_1803bc1e0);
                                                          uVar8 = 0x174;
                                                          goto LAB_1801ce03f;
                                                        }
LAB_1801cdfb7:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        uVar8 = 0x174;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#113="style_guide_block" */
                                                                            DAT_1803bc198,uVar8);
                                                      if (cVar3 == '\0') goto LAB_1801cdfb7;
/* novels.prompts.inkos.reviser_prompts#114="length_guidance_block" */
                                                      FUN_18009b380(plVar11,DAT_1803bc1a0,
/* novels.prompts.inkos.reviser_prompts#3="" */
                                                                    DAT_1803bbe28);
/* novels.prompts.inkos.reviser_prompts#188="str" */
                                                      uVar8 = DAT_1803bc3f0;
                                                      lVar12 = FUN_18036f670(param_1,plVar11,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                             DAT_1803bc1e0);
                                                      if (lVar12 == 0) {
                                                        cVar3 = FUN_18009b370();
                                                        if (cVar3 == '\0') goto LAB_1801ce035;
                                                        FUN_180373340(param_1,&local_98,
/* novels.prompts.inkos.reviser_prompts#122="__annotations__" */
                                                                      DAT_1803bc1e0);
LAB_1801ce03a:
                                                        uVar8 = 0x175;
                                                        goto LAB_1801ce03f;
                                                      }
                                                      cVar3 = FUN_18000d470(param_1,lVar12,
/* novels.prompts.inkos.reviser_prompts#114="length_guidance_block" */
                                                                            DAT_1803bc1a0,uVar8);
                                                      if (cVar3 == '\0') {
LAB_1801ce035:
                                                        FUN_18000ce30(param_1,&local_98);
                                                        goto LAB_1801ce03a;
                                                      }
                                                      FUN_180001d80(param_1);
/* novels.prompts.inkos.reviser_prompts#185="ReviserUserPromptParams" */
                                                      piVar1 = DAT_1803bc3d8;
                                                      local_90 = PyType_Type_exref;
                                                      plVar13 = (longlong *)FUN_180374a40(param_1,3)
                                                      ;
                                                      if (*piVar1 + 1 != 0) {
                                                        *piVar1 = *piVar1 + 1;
                                                      }
                                                      FUN_18000cc80(plVar13,0,piVar1);
                                                      if (*DAT_1803ab240 + 1 != 0) {
                                                        *DAT_1803ab240 = *DAT_1803ab240 + 1;
                                                      }
                                                      FUN_18000cc80(plVar13,1);
                                                      iVar4 = (int)*plVar11 + 1;
                                                      if (iVar4 != 0) {
                                                        *(int *)plVar11 = iVar4;
                                                      }
                                                      FUN_18000cc80(plVar13,2,plVar11);
                                                      *(undefined4 *)(lVar7 + 0x28) = 0x167;
                                                      plVar14 = (longlong *)
                                                                FUN_180001790(param_1,local_90,
                                                                              plVar13,plVar9);
                                                      if ((-1 < (int)*plVar13) &&
                                                         (*plVar13 = *plVar13 + -1, *plVar13 == 0))
                                                      {
                                                        lVar10 = FUN_18000cc40(plVar13);
                                                        (**(code **)(lVar10 + 0x30))(plVar13);
                                                      }
                                                      if (plVar14 != (longlong *)0x0) {
                                                        iVar4 = (int)*plVar14 + 1;
                                                        if (iVar4 != 0) {
                                                          *(int *)plVar14 = iVar4;
                                                        }
                                                        if ((-1 < (int)*plVar11) &&
                                                           (*plVar11 = *plVar11 + -1, *plVar11 == 0)
                                                           ) {
                                                          lVar10 = FUN_18000cc40(plVar11);
                                                          (**(code **)(lVar10 + 0x30))(plVar11);
                                                        }
                                                        if ((-1 < (int)*plVar14) &&
                                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)
                                                           ) {
                                                          lVar10 = FUN_18000cc40(plVar14);
                                                          (**(code **)(lVar10 + 0x30))(plVar14);
                                                        }
                                                        plVar13 = local_88;
                                                        *(undefined4 *)(lVar7 + 0x28) = 0x166;
                                                        lVar10 = FUN_180376770(param_1,local_88,
                                                                               plVar14);
                                                        if ((-1 < (int)*plVar13) &&
                                                           (*plVar13 = *plVar13 + -1, *plVar13 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar13);
                                                          (**(code **)(lVar12 + 0x30))(plVar13);
                                                        }
                                                        if ((-1 < (int)*plVar14) &&
                                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)
                                                           ) {
                                                          lVar12 = FUN_18000cc40(plVar14);
                                                          (**(code **)(lVar12 + 0x30))(plVar14);
                                                        }
                                                        if (lVar10 != 0) {
/* novels.prompts.inkos.reviser_prompts#185="ReviserUserPromptParams" */
                                                          FUN_1801c0a10(DAT_1803ca608,DAT_1803bc3d8,
                                                                        lVar10);
                                                          FUN_180001d80(param_1);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar7 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar7 + 0x30))(plVar9);
                                                          }
                                                          if ((-1 < (int)*plVar11) &&
                                                             (*plVar11 = *plVar11 + -1,
                                                             *plVar11 == 0)) {
                                                            lVar7 = FUN_18000cc40(plVar11);
                                                            (**(code **)(lVar7 + 0x30))(plVar11);
                                                          }
                                                          uVar8 = FUN_18036fb80(param_1,
                                                  DAT_1803bc3f8);
                                                  uVar8 = FUN_1801cb120(param_1,uVar8);
/* novels.prompts.inkos.reviser_prompts#190="build_reviser_user_prompt" */
                                                  FUN_1801c0a10(DAT_1803ca608,DAT_1803bc400,uVar8);
                                                  piVar1 = DAT_1803ca610;
                                                  if (*DAT_1803ca610 + 1 == 0) {
                                                    return DAT_1803ca610;
                                                  }
                                                  *DAT_1803ca610 = *DAT_1803ca610 + 1;
                                                  return piVar1;
                                                  }
                                                  goto LAB_1801ce23d;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  lVar10 = local_98;
                                                  FUN_18000ce70(&local_98);
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar12 + 0x30))(plVar11);
                                                  }
                                                  local_98 = lVar10;
                                                  FUN_18000ce70(&local_98);
                                                  uVar8 = 0x167;
                                                  local_98 = lVar10;
                                                  }
                                                  lVar10 = local_98;
                                                  FUN_18000ce70(&local_98);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar12 + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar12 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar12 + 0x30))(plVar11);
                                                  }
                                                  goto LAB_1801ce29b;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  }
                                                  }
                                                  uVar8 = 0x3a;
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x39;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x38;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x37;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x36;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x35;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x34;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x33;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x32;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x31;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cd698;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x30;
                                                  }
                                                  }
LAB_1801cd698:
                                                  lVar10 = local_98;
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar10 + 0x30))(plVar9);
                                                    lVar10 = local_98;
                                                  }
                                                  goto LAB_1801ce29b;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar10 + 0x30))(plVar9);
                                                  }
                                                  uVar8 = 0x2f;
                                                  lVar10 = local_98;
                                                  }
                                                  }
                                                  goto LAB_1801ce29b;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  }
                                                  }
                                                  uVar8 = 0x2b;
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x2a;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x29;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x28;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x27;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x26;
                                                  }
                                                  }
                                                  }
                                                  goto LAB_1801cc71e;
                                                  }
                                                  FUN_18000ce30(param_1,&local_98);
                                                  }
                                                  if ((-1 < (int)*plVar11) &&
                                                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                    lVar10 = FUN_18000cc40(plVar11);
                                                    (**(code **)(lVar10 + 0x30))(plVar11);
                                                  }
                                                  uVar8 = 0x25;
                                                  }
                                                }
                                              }
                                              goto LAB_1801cc71e;
                                            }
                                            FUN_18000ce30(param_1,&local_98);
                                          }
                                          if ((-1 < (int)*plVar11) &&
                                             (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                            lVar10 = FUN_18000cc40(plVar11);
                                            (**(code **)(lVar10 + 0x30))(plVar11);
                                          }
                                          uVar8 = 0x24;
                                        }
                                      }
LAB_1801cc71e:
                                      lVar10 = local_98;
                                      if ((-1 < (int)*plVar9) &&
                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                        lVar10 = FUN_18000cc40(plVar9);
                                        (**(code **)(lVar10 + 0x30))(plVar9);
                                        lVar10 = local_98;
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
LAB_1801ce29b:
  local_98 = lVar10;
  piVar1 = *(int **)(local_98 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar7 = FUN_18038e2a0(lVar7,uVar8);
  }
  else {
    if (*(longlong *)(piVar1 + 6) == lVar7) goto LAB_1801ce2dd;
    lVar7 = FUN_18038e2a0(lVar7,uVar8);
    *(int **)(lVar7 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_98,lVar7);
LAB_1801ce2dd:
  FUN_180001d80(param_1);
  FUN_18000ce40(param_1,&local_98);
  return (int *)0x0;
}
